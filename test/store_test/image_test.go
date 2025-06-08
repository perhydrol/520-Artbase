package store

import (
	"context"
	"crypto/sha256"
	"demo520/internal/520/store"
	"demo520/internal/pkg/model"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"testing"
)

var userCount = 3

func setupDatabase() (*gorm.DB, []model.UserM, error) {
	// 3. 构造 DSN
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local", "testpassword")

	// 4. 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.UserM{}); err != nil {
		return nil, nil, err
	}
	if err := db.AutoMigrate(&model.ImageM{}); err != nil {
		return nil, nil, err
	}
	if err := db.AutoMigrate(&model.ImageTagM{}); err != nil {
		return nil, nil, err
	}

	userStore := store.NewStore(db).User()
	ctx := context.Background()
	users := make([]model.UserM, userCount)

	for i := range users {
		users[i].UserUUID = uuid.New().String()
		users[i].Email = faker.Email()
		users[i].Password = faker.Password()
		users[i].Nickname = faker.Name()
	}
	for i := range users {
		if err := userStore.Create(ctx, &users[i]); err != nil {
			return nil, nil, err
		}
	}

	return db, users, nil
}

func TestImageStore(t *testing.T) {
	db, users, err := setupDatabase()
	if err != nil {
		t.Fatal(err)
	}

	imageStore := store.NewStore(db).Image()
	ctx := context.Background()
	imagesCount := 20
	images := make([]model.ImageM, imagesCount*userCount)
	t.Run("Create", func(t *testing.T) {
		wg := sync.WaitGroup{}
		for userIndex, user := range users {
			for i := userIndex * imagesCount; i < imagesCount*(userIndex+1); i++ {
				wg.Add(1)
				go func(i int, user model.UserM) {
					defer wg.Done()
					images[i].ImageUUID = faker.UUIDHyphenated()
					images[i].UserUUID = user.UserUUID
					images[i].Token = faker.UUIDHyphenated()
					hasher := sha256.New()
					if _, err := io.Copy(hasher, strings.NewReader(images[i].ImageUUID)); err != nil {
						t.Error(err)
						return
					}
					images[i].Hash = fmt.Sprintf("%x", hasher.Sum(nil))
					images[i].IsPublic = i < (imagesCount-1)%2
					images[i].Tags = []model.ImageTagM{
						{Tag: faker.Word(), ImageUUID: images[i].ImageUUID},
						{Tag: faker.Word(), ImageUUID: images[i].ImageUUID},
						{Tag: faker.Word(), ImageUUID: images[i].ImageUUID},
					}
					require.NoError(t, imageStore.Create(ctx, &images[i]))
					getImg, err := imageStore.Get(ctx, images[i].ImageUUID)
					require.NoError(t, err)
					assert.True(t, getImg.Equal(&images[i]))
				}(i, user)
			}
		}
		wg.Wait()
	})

	t.Run("add tags ti images", func(t *testing.T) {
		newTags := make([]string, 2)
		for i := range newTags {
			newTags[i] = faker.Word()
		}
		wg := sync.WaitGroup{}
		for index := range images {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if err := imageStore.AddTagsToImage(ctx, images[i].ImageUUID, newTags); err != nil {
					t.Errorf("failed to add tags to image %d: %v", i, err)
					return
				}
				getImage, err := imageStore.Get(ctx, images[i].ImageUUID)
				if err != nil {
					t.Errorf("failed to get image %d after adding tags: %v", i, err)
					return
				}
				count := 0
				for _, value := range getImage.Tags {
					if slices.Contains(newTags, value.Tag) {
						count++
					}
				}
				assert.Equal(t, len(newTags), count)
			}(index)
		}
		wg.Wait()
	})

	t.Run("delete tags images", func(t *testing.T) {
		wg := sync.WaitGroup{}
		for index := range images {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				getImage, err := imageStore.Get(ctx, images[i].ImageUUID)
				if err != nil {
					t.Errorf("failed to get image %d for tag deletion: %v", i, err)
					return
				}
				index := rand.Intn(len(getImage.Tags))
				delTags := []string{getImage.Tags[index].Tag}
				if err := imageStore.DeleteTagFromImage(ctx, getImage.ImageUUID, delTags); err != nil {
					t.Errorf("failed to delete tag from image %d: %v", i, err)
					return
				}
				getImage, err = imageStore.Get(ctx, images[i].ImageUUID)
				if err != nil {
					t.Errorf("failed to get image %d after deleting tag: %v", i, err)
					return
				}
				for _, value := range getImage.Tags {
					if value.Tag == delTags[0] {
						t.Errorf("tag %q was not deleted", delTags[0])
					}
				}
			}(index)
		}
		wg.Wait()
	})

	var allImagesUUID []string
	t.Run("Get User Images", func(t *testing.T) {
		for _, user := range users {
			imageVal, images, err := imageStore.GetUserImages(ctx, user.UserUUID, 0, imagesCount)
			if err != nil {
				t.Fatalf("failed to get User Images: %v", err)
			}
			assert.Equal(t, len(images), imagesCount)
			assert.Equal(t, imageVal, int64(imagesCount))
			for _, image := range images {
				allImagesUUID = append(allImagesUUID, image.ImageUUID)
			}
		}
	})

	t.Run("Get random Public Images", func(t *testing.T) {
		imageVal, images, err := imageStore.GetRandomPublicImages(ctx, imagesCount)
		if err != nil {
			t.Fatalf("failed to get random User Images: %v", err)
		}
		assert.True(t, len(images) == imageVal)
		for _, image := range images {
			assert.True(t, slices.Contains(allImagesUUID, image.ImageUUID))
		}
	})

	t.Run("Delete Images", func(t *testing.T) {
		wg := sync.WaitGroup{}
		for _, imageUUID := range allImagesUUID {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := imageStore.Delete(ctx, imageUUID); err != nil {
					t.Errorf("failed to delete image %s: %v", imageUUID, err)
					return
				}
			}()
		}
		wg.Wait()
	})

	t.Run("Get User Images after del", func(t *testing.T) {
		for _, user := range users {
			imageVal, images, err := imageStore.GetUserImages(ctx, user.UserUUID, 0, 10)
			if err != nil {
				t.Fatalf("failed to get User Images: %v", err)
			}
			assert.True(t, len(images) == 0)
			assert.Equal(t, imageVal, int64(0))
		}
	})

	t.Run("Get random Public Images after del", func(t *testing.T) {
		imageVal, images, err := imageStore.GetRandomPublicImages(ctx, 10)
		if err != nil {
			t.Fatalf("failed to get random User Images: %v", err)
		}
		assert.True(t, len(images) == imageVal)
		assert.Equal(t, imageVal, 0)
	})
}
