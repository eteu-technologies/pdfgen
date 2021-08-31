# pdfgen

## Running

```bash
docker run --rm -ti \
    --name pdfgen \
    --read-only \
    --tmpfs /tmp \
    --shm-size 2gb \
    -p 5000:5000 \
    eteu/pdfgen
```

## License

LGPL 3.0
