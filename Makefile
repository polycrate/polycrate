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

unexport GITHUB_TOKEN
release: latest
	git push origin main
	goreleaser release --rm-dist --debug
	
	
latest: tag
	echo "$(shell svu --strip-prefix)" > latest
	cat latest
	mc cp latest ayedo-s3/polycrate/cli
	rm latest

check:
	goreleaser check

serve:
	mkdocs serve

changelog:
	git-chglog --next-tag $(shell svu next) --output docs/changelog/$(shell svu next).md $(shell svu next)
	git add .
	git commit -am "changelog created $(shell svu next)"