package errno

import "net/http"

var ErrImageNotFound = &Errno{HTTP: 404, Code: "ResourceNotFound.ImageNotFound", Message: "Image not found"}
var ErrImageJSONNotFound = &Errno{HTTP: http.StatusNotFound, Code: "ResourceNotFound.ImageJSON", Message: "Image metadata JSON not found"}
var ErrImageJSONInvalid = &Errno{HTTP: http.StatusBadRequest, Code: "InvalidParameter.ImageJSON", Message: "Invalid image metadata JSON format"}
var ErrImageFileInvalid = &Errno{
	HTTP:    http.StatusBadRequest,
	Code:    "InvalidParameter.ImageFileInvalid",
	Message: "Invalid image file format or content",
}
var ErrImageFileTooLarge = &Errno{
	HTTP:    http.StatusRequestEntityTooLarge,
	Code:    "LimitExceeded.ImageFileTooLarge",
	Message: "Image file size exceeds the maximum allowed limit",
}
