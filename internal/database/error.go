package database

import "errors"

var NotFound = errors.New("entity not found")
var Duplicate = errors.New("duplicate entity")

