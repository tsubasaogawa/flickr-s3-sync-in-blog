# flickr-s3-sync-in-blog (fssync)

flickr-s3-sync-in-blog helps you migrate from Flickr.

## Features

- Copy images in the blog post from Flickr to S3
- Replace Flickr URLs with S3 URLs in the blog post
- Remove some tags and attributes used by Flickr, such as `data-flickr-embed`, `client-code.js`

### For example

Before 

```html
<!-- /home/ipsum/entry/2024/04/24/123456.md -->
<a data-flickr-embed="true" data-header="true" title="twilight" href="https://www.flickr.com/photos/tsubasaogawa/123456789/in/dateposted-public/">
    <img src="https://c1.staticflickr.com/1/234/56789_123abc456def_c.jpg" alt="twilight" width="530" height="800" />
</a>
<script async src="//embedr.flickr.com/assets/client-code.js" charset="utf-8"></script>
```

Run `fssync`

```bash
./fssync -config=fssync.toml /home/ipsum/entry/2024/04/24/123456.md

2024/04/24 22:16:26 Load: /home/ipsum/entry/2024/04/24/123456.md
2024/04/24 22:16:27 Upload: 56789_123abc456def_c.jpg
2024/04/24 22:16:27 Backup: /home/ipsum/blog_backups/2024/04/24/123456.md.bak
2024/04/24 22:16:27 Save: /home/ipsum/entry/2024/04/24/123456.md
```

After

```html
<a tabindex="-1">
    <img src="https://your.s3bucket.com/56789_123abc456def_c.jpg" alt="twilight" width="530" height="800" />
</a>
```

```bash
aws s3 ls s3://your.s3bucket.com/56789_123abc456def_c.jpg
# -> File exists
```

## Install

Download the binary from the [release](https://github.com/tsubasaogawa/flickr-s3-sync-in-blog/releases) page.

## Configuration

Open and edit `fssync.toml`.

## Usage

```bash
./fssync -config=fssync.toml -dryrun=true <PATH TO BLOG FILE>
```

`fssync` with dryrun mode will not upload an image or modify a blog file. It outputs replaced blog texts to stdout.

After your check, you can run it without dryrun option.

```bash
./fssync -config=fssync.toml <PATH TO BLOG FILE>
```

## Tips

### Run fssync to multiple blog posts

`find` is useful.

```bash
find <PATH TO BLOG DIR> -name '*.md' -type f -print0 | xargs -0 -n1 ./fssync -config=fssync.toml
```

### For hatenablog using [blogsync](https://github.com/x-motemen/blogsync)

```bash
# Pull all blog posts
blogsync pull <blogID>

# Run fssync with dryrun mode
cd <PATH TO fssync DIR>
./fssync -config=fssync.toml -dryrun=true <PATH TO BLOG FILE>

# Run fssync for all posts in 2024
find <PATH TO BLOG DIR> -wholename '**/2024/*.md' -type f -print0 | xargs -0 -n1 ./fssync -config=fssync.toml

# Push modified blog posts (in 5 minutes)
find <PATH TO BLOG DIR> -name '*.md' -mmin -5 -type f -print0 | xargs -0 -n1 blogsync push
```
