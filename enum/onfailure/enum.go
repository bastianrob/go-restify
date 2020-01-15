package onfailure

import "github.com/SpaceStock/go-restify/enum"

//Enumeration constants
const (
	Exit        = enum.OnFailure("exit")
	Fallthrough = enum.OnFailure("fallthrough")
)
