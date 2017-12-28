DEBUG=false
CC=gcc
LDFLAGS=''
EXTLDFLAGS=''
OS=linux
ARCH=amd64
ENGAGEMENT_TIME=0
BINDSTRING=0.0.0.0:4444
CLI_KF=id-cli
SRV_KF=id-srv
GO_EXT_VARS="-X main.bdPriv=$(shell cat $(SRV_KF)) -X main.bdPub=$(shell cat $(SRV_KF).pub) -X main.cliPub=$(shell cat $(CLI_KF).pub) -X main.Debug=$(DEBUG) -X main.killTime=$(ENGAGEMENT_TIME) -X main.Method=$(METHOD) -X main.BindString=$(BINDSTRING)"
