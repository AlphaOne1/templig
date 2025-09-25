# SPDX-FileCopyrightText: 2025 The templig contributors.
# SPDX-License-Identifier: MPL-2.0

.PHONY: all clean run

all: config.go

config.go: schema.json
	go tool go-jsonschema	\
		$<					\
		--only-models		\
		--schema-package=https://github.com/AlphaOne1/templig/examples/configSchema/schema=main	\
		--schema-output=https://github.com/AlphaOne1/templig/examples/configSchema/schema=$@	\
		--schema-root-type=https://github.com/AlphaOne1/templig/examples/configSchema/schema=Config

run: main.go config.go
	go run $^

clean:
	-rm -f config.go