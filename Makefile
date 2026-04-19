# 标准入口：与 scripts/run-tests.sh、CI 一致
.PHONY: test verify docs-check fmt vet check docker-build

test:
	bash scripts/run-tests.sh

# 仅格式化（修改文件）
fmt:
	go fmt ./...

vet:
	go vet ./...

# fmt 检查 + vet + test（提交前建议执行）
check:
	bash scripts/check-go.sh

verify: docs-check

docs-check:
	bash scripts/verify-docs.sh

docker-build:
	docker build -t ase:local .
