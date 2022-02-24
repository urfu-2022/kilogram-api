#!/bin/bash

GOARCH=386 GOOS=linux go build .
ssh dimastark@84.201.142.5 'sudo service kilogram-api stop'
scp kilogram-api dimastark@84.201.142.5:/home/dimastark
ssh dimastark@84.201.142.5 'sudo service kilogram-api start'
