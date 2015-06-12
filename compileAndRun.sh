#!/bin/sh
#gom install
gom exec go install -race .
communitrix -logLevel=DEBUG
