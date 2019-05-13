#!/bin/sh

build() {
	path=main.go
	# build
	GOOS=linux go build -o $1 $path
	zip -j $1.zip $1
	chmod 777 $1.zip
}

build jmeme
cd terraform && terraform taint aws_lambda_alias.jmeme_live && terraform taint aws_lambda_function.jmeme
