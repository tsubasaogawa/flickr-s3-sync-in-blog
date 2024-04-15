# flickr-s3-sync-from-blog

flickr-s3-sync-from-blog is an image transfer embedded in blog posts from Flickr to Amazon S3.

## Tips

### Convert all entries

```bash
 find /mnt/c/Users/Tsubasa\ Ogawa/Documents/blogsync -wholename '**/2015*.md' -type f -print0 | xargs 
-0 -n1 go run */**.go --s3Bucket=photo.ogatube.com --s3Dir=blog --uploadS3=true --overwrite=true --backupDir=/var/tmp/ogawa/tsubasa/
```

### List up entries which is fixed flickr url by converter

```bash
find /mnt/c/Users/Tsubasa\ Ogawa/Documents/blogsync/ -name '*.md' -mtime -1 -type f
```
