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
	echo ${AYEDO_CARGO_PASSWORD} | docker login cargo.ayedo.cloud -u ${AYEDO_CARGO_USERNAME} --password-stdin

snapshot:
	goreleaser release --snapshot --rm-dist --debug --timeout 120m

unexport GITLAB_TOKEN
build:

unexport GITLAB_TOKEN
release: latest
	git push origin main
	goreleaser release --rm-dist --debug --timeout 120m
	
latest: tag
	echo "$(shell svu --strip-prefix)" > latest
	cat latest
	mc cp latest ayedo-s3/polycrate/cli
	rm latest

polyhub:
	echo "Uploading linux 386 package to Polyhub"
	curl --location --request POST 'https://hub.polycrate.io/api/v1/file/upload/polycrate/linux/386/polycrate_$(shell svu --strip-prefix)_linux_386.tar.gz/$(shell svu --strip-prefix)' \
	--header 'Authorization: Bearer ${POLYHUB_TOKEN}' \
	--header 'Content-Type: application/gzip' \
	--data-binary '@./dist/polycrate_$(shell svu --strip-prefix)_linux_386.tar.gz'
	echo "Uploading linux amd64 package to Polyhub"
	curl --location --request POST 'https://hub.polycrate.io/api/v1/file/upload/polycrate/linux/amd64/polycrate_$(shell svu --strip-prefix)_linux_amd64.tar.gz/$(shell svu --strip-prefix)' \
	--header 'Authorization: Bearer ${POLYHUB_TOKEN}' \
	--header 'Content-Type: application/gzip' \
	--data-binary '@./dist/polycrate_$(shell svu --strip-prefix)_linux_amd64.tar.gz'
	echo "Uploading linux arm64 package to Polyhub"
	curl --location --request POST 'https://hub.polycrate.io/api/v1/file/upload/polycrate/linux/arm64/polycrate_$(shell svu --strip-prefix)_linux_arm64.tar.gz/$(shell svu --strip-prefix)' \
	--header 'Authorization: Bearer ${POLYHUB_TOKEN}' \
	--header 'Content-Type: application/gzip' \
	--data-binary '@./dist/polycrate_$(shell svu --strip-prefix)_linux_arm64.tar.gz'
	echo "Uploading darwin amd64 package to Polyhub"
	curl --location --request POST 'https://hub.polycrate.io/api/v1/file/upload/polycrate/darwin/amd64/polycrate_$(shell svu --strip-prefix)_darwin_amd64.tar.gz/$(shell svu --strip-prefix)' \
	--header 'Authorization: Bearer ${POLYHUB_TOKEN}' \
	--header 'Content-Type: application/gzip' \
	--data-binary '@./dist/polycrate_$(shell svu --strip-prefix)_darwin_amd64.tar.gz'
	echo "Uploading darwin arm64 package to Polyhub"
	curl --location --request POST 'https://hub.polycrate.io/api/v1/file/upload/polycrate/darwin/arm64/polycrate_$(shell svu --strip-prefix)_darwin_arm64.tar.gz/$(shell svu --strip-prefix)' \
	--header 'Authorization: Bearer ${POLYHUB_TOKEN}' \
	--header 'Content-Type: application/gzip' \
	--data-binary '@./dist/polycrate_$(shell svu --strip-prefix)_darwin_arm64.tar.gz'

check:
	goreleaser check

serve:
	mkdocs serve

# changelog:
# 	git-chglog --next-tag $(shell svu next) --output docs/changelog/$(shell svu next).md $(shell svu next)
# 	git add .
# 	git commit -am "changelog created $(shell svu next)"