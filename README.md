# nuc-chain

```bash
rm -rf ~/.nuc/
mkdir ~/.nuc/
docker run --name nucnode --restart=always -p 7121:7121 -p 17001:17001 \ 
-v ~/.nuc/:/root/.nuc -it -d 103.28.213.52:99/nuc/node:v19
```
