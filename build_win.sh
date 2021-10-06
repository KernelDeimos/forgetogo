# sudo apt-get install gcc-multilib
# sudo apt-get install gcc-mingw-w64

cd src
mkdir -p ../builds/win
GOOS=windows GOARCH=amd64 go build -o ../builds/win/ftg.exe ./forgetogo/main.go

