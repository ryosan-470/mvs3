PROJECT = move-to-s3
fmt:
	go fmt

build:
	GOARCH=amd64 GOOS=linux go build -o build/$(PROJECT)
.PHONY: build

deploy:
	aws cloudformation package \
			--template-file template.yaml \
			--s3-bucket $(PROJECT) \
			--output-template-file .template-output.yaml

	aws cloudformation deploy \
			--template-file .template-output.yaml \
			--stack-name $(PROJECT) \
			--capabilities CAPABILITY_IAM \
			--parameter-overrides \
					TopicArn=${TOPICARN} \
					TargetFileName=${TARGET_FILENAME} \
					OriginBucket=${ORIGIN_BUCKET} \
					OriginRegion=${ORIGIN_REGION} \
					TargetBucket=${TARGET_BUCKET} \
					TargetRegion=${TARGET_REGION} \
					Password=${PASSWORD}

init:
	aws s3 mb s3://$(PROJECT)

destroy:
	aws cloudformation delete-stack --stack-name $(PROJECT)
	aws s3 rb --force s3://$(PROJECT)
