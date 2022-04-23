tag:
	git tag $(svu next)

next:
	svu next

docker-login:
	echo $GH_TOKEN | docker login ghcr.io -u derfabianpeter --password-stdin