package errorset

import "errors"

var ErrInternalServer = errors.New("internal server error")
var ErrDatabase = errors.New("database error")
var ErrOAuth = errors.New("oauth error")
var ErrSQL = errors.New("sql error")
var ErrForbidden = errors.New("Forbidden")
var ErrParams = errors.New("payload error")
