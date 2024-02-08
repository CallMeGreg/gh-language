git tag v1.0.0
git push origin v1.0.0
GOOS=windows GOARCH=amd64 go build -o gh-language-windows-amd64.exe
GOOS=linux GOARCH=amd64 go build -o gh-language-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o gh-language-darwin-amd64
gh release create v1.0.0 ./*amd64*