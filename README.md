# flickr-s3-sync-from-blog

flickr-s3-sync-from-blog is an image transfer embedded in a blog post from Flickr to Amazon S3.

## Features

- Copy image from Flickr to S3 based on the blog post file
- Replace Flickr URLs with S3 URLs in the file
- Remove some tags or attributes used by Flickr such as `data-flickr-embed`, `client-code.js`

For example:

Before 

```html
<a data-flickr-embed="true" data-header="true" title="twilight" href="https://www.flickr.com/photos/tsubasaogawa/123456789/in/dateposted-public/"><img src="https://c1.staticflickr.com/1/234/56789_123abc456def_c.jpg" alt="twilight" width="530" height="800" /></a><script async src="//embedr.flickr.com/assets/client-code.js" charset="utf-8"></script>
```

After

```html
<a tabindex="-1"><img src="https://photo.ogatube.com/blog/56789_123abc456def_c_c.jpg" alt="twilight" width="530" height="800" /></a>
```

## Install

Download the binary from the [release](https://github.com/tsubasaogawa/flickr-s3-sync-from-blog/releases) page.

## Configuration

Open and edit `fssync.toml`.

## Usage

```bash
./fssync -config=fssync.toml -dryrun=true <PATH TO BLOG FILE>
```

fssync will not upload an image or overwrite a file. It outputs the replaced blog text to stdout.

After your check, you can run it with no dryrun option.

## Tips

### Use fssync for hatenablog using [blogsync](https://github.com/x-motemen/blogsync)

```bash
# Pull all blog posts
blogsync pull <blogID>

# Run fssync with dryrun mode
cd <PATH TO fssync DIR>
./fssync -config=fssync.toml -dryrun=true <PATH TO BLOG FILE>

# Run fssync for all posts
find <PATH TO BLOG DIR> -wholename '**/*.md' -type f -print0 | xargs -0 -n1 ./fssync -config=fssync.toml

# Push modified blog posts (in 5 minutes)
find <PATH TO BLOG DIR> -name '*.md' -mmin -5 -type f -print0 | xargs -0 -n1 blogsync push
```
