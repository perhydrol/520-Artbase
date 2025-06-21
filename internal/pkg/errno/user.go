package errno

var (
	// ErrUserAlreadyExist 代表用户已经存在.
	ErrUserAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.UserAlreadyExist", Message: "User already exist."}

	ErrUserLoginRequestOutTime = &Errno{HTTP: 401, Code: "FailedOperation.UserLoginRequestOutTime", Message: "Login request failed."}

	// ErrUserNotFound 表示未找到用户.
	ErrUserNotFound = &Errno{HTTP: 401, Code: "ResourceNotFound.UserNotFoundOrPasswordIncorrect", Message: "Incorrect username or password"}

	// ErrPasswordIncorrect 表示密码不正确.
	ErrPasswordIncorrect = &Errno{HTTP: 401, Code: "ResourceNotFound.UserNotFoundOrPasswordIncorrect", Message: "Incorrect username or password"}
)
