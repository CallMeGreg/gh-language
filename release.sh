export version=v2.0.2
git tag $version
git push origin $version
GOOS=windows GOARCH=amd64 go build -o gh-language-windows-amd64.exe
GOOS=linux GOARCH=amd64 go build -o gh-language-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o gh-language-darwin-amd64
gh release create $version ./*amd64*