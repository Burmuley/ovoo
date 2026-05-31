package config

type CfgSectionName string

const (
	APISection       CfgSectionName = "api"
	MilterSection    CfgSectionName = "milter"
	SocketMapSection CfgSectionName = "socketmap"
)
