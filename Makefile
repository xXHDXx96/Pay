.PHONY: build build-linux build-android build-darwin build-windows frontend frontend-build frontend-dev clean package

build: build-linux

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/alipay-payment-linux-amd64 ./cmd/server/
	GOOS=linux GOARCH=arm64 go build -o bin/alipay-payment-linux-arm64 ./cmd/server/

build-android:
	GOOS=linux GOARCH=arm64 go build -o bin/alipay-payment-android-arm64 ./cmd/server/
	GOOS=linux GOARCH=arm go build -o bin/alipay-payment-android-arm ./cmd/server/

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/alipay-payment-darwin-amd64 ./cmd/server/
	# Note: darwin/arm64 (Apple Silicon) requires Go 1.19+ due to Go 1.18 compiler bug
	# GOOS=darwin GOARCH=arm64 go build -o bin/alipay-payment-darwin-arm64 ./cmd/server/

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/alipay-payment-windows-amd64.exe ./cmd/server/

frontend: frontend-build

frontend-build:
	cd web && npm run build

frontend-dev:
	cd web && npm run dev

package: build-linux frontend-build
	mkdir -p dist
	cp bin/alipay-payment-linux-amd64 dist/alipay-payment
	cp -r web/dist dist/frontend
	tar czf alipay-payment-app.tar.gz -C dist .

clean:
	rm -rf bin/ web/dist/ dist/
