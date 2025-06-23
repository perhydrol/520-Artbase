package store_test

import (
	"demo520/test/testhelper"
	"testing"
)

// setupStoreTest 设置store层测试环境
func setupStoreTest(t *testing.T) *testhelper.TestSuite {
	// 使用默认的测试数据库配置
	config := testhelper.DefaultTestDBConfig()

	// 创建测试套件
	ts := testhelper.NewTestSuite(t, config)

	// 清理数据库
	ts.Cleanup(t)

	return ts
}

// teardownStoreTest 清理store层测试环境
func teardownStoreTest(t *testing.T, ts *testhelper.TestSuite) {
	ts.Cleanup(t)
}
