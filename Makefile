tag:
	git tag $(shell svu next)
	git push origin $(shell svu)

next:
	svu next

docker-login:
	echo ${GITHUB_TOKEN} | docker login ghcr.io -u ${GHCR_USER} --password-stdin

snapshot:
	goreleaser release --snapshot --rm-dist --debug

release:
	goreleaser release

check:
	goreleaser check