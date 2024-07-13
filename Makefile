fmt:
	echo "Formatting lua/yankbank..."
	stylua lua/ --config-path=.stylua.toml
	echo "Formatting Go files in ./remote..."
	find ./remote -name '*.go' -exec gofmt -w {} +

lint:
	echo "Linting lua/yankbank..."
	luacheck lua/ --globals vim

pr-ready: fmt lint
