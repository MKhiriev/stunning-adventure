hey -m POST -n 2000 -c 50 -disable-compression \
  -D profiles/post_updates_body.gz \
  -H "Accept: application/json" \
  -H "Content-Encoding: gzip" \
  -H "Content-Type: application/json" \
  -H "Hashsha256: 00a74b9ead06b09fd3683158b77529b63507665ed218d9097fd8d768ab63ded8" \
  http://localhost:8081/updates/