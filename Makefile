tag:
	git tag $(shell svu next)
	git push origin $(shell svu)

next:
	svu next

docker-login:
	echo ${GH_TOKEN} | docker login ghcr.io -u derfabianpeter --password-stdin

snapshot:
	goreleaser release --snapshot --rm-dist --debug

release:
	goreleaser release