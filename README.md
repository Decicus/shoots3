# shoots3

Simple commandline tool to store images in S3.

![](https://media.giphy.com/media/l4JzgUo64J3DaxnI4/giphy.gif)

## Installation

```sh
go get github.com/nerdenough/shoots3
```

Make sure your AWS environment variables are set correctly:

```sh
AWS_ACCESS_KEY_ID=[access key id]
AWS_SECRET_ACCESS_KEY=[secret access key]
AWS_REGION=[same region as your bucket]
```

Set an environment variable if you want to use a bucket as default:

```sh
SHOOTS3_DEFAULT_BUCKET=my-bucket-name
```

## Example

```sh
shoots3 image.png
```

with [aws-vault](https://github.com/99designs/aws-vault)

```sh
aws-vault exec user -- shoots3 image.png
```

## Usage

```sh
shoots3 [flags] filename
```

### Flags
| Flag | Default                      | Usage                                |
| ---- | ---------------------------- | ------------------------------------ |
| `-k` | random string                | Custom key to store the image        |
| `-l` | 6                            | Length of the randomly generated url |
| `-b` | `env.SHOOTS3_DEFAULT_BUCKET` | Bucket to upload the image to        |
| `-r` | `env.AWS_REGION`             | Region where the bucket is located   |
