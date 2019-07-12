package onfailure

import "github.com/bastianrob/go-restify/enum"

//Enumeration constants
const (
	Exit        = enum.OnFailure("exit")
	Fallthrough = enum.OnFailure("fallthrough")
)
