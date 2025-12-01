@echo off
go build -o server.exe cmd/api/main.go
if %errorlevel% equ 0 (
    ./server.exe
)