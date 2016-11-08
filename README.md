
### Starting dev server
``` bash
~/bin/go_appengine/dev_appserver.py .
```

### Uploading prod

```
 ~/bin/go_appengine/appcfg.py update .

```


### Populating data

```
curl -X POST -d "{\"markerName\":\"SNP_A-1794291\",\"rsID\":\"\",\"chromosome\":\"9\",\"position\":74393843,\"a_allele\":\"\",\"b_allele\":\"\",\"arrays\":[\"Affymetrix SNP 6.0\"]}" "http://markermaker-148719.appspot.com/process/"
```



### Retrieving data
See it in a browser 

http://markermaker-148719.appspot.com/markerquery/SNP_A-1794291


Or curl it and get json

```
curl "http://markermaker-148719.appspot.com/markerqueryraw/SNP_A-1794291"
```
