@echo off
echo Running LearnVibe unit tests...

cd %~dp0..
echo Testing models...
cd cms && go test -v ./models

echo Testing middleware...
go test -v ./middleware

echo All unit tests completed. 