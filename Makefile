tag:
	git tag $(shell svu next)
	git push origin $(shell svu)
	#echo $(shell svu) > latest.txt

delete-tag:
	git tag -d $(shell svu)
	git push --delete origin $(shell svu)

next:
	@svu next

docker-login:
	echo ${GITHUB_TOKEN} | docker login ghcr.io -u ${GHCR_USER} --password-stdin

snapshot:
	goreleaser release --snapshot --rm-dist --debug

release: tag
	git push origin main
	goreleaser release --rm-dist --debug

check:
	goreleaser check