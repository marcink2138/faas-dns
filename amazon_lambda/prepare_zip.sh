#/bin/bash
rm main
rm main.zip
echo "Compiling function"
GOOS=linux GOARCH=amd64 go build -o main function.go
echo "Creating zip"
zip main.zip main

