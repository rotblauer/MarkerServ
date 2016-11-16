

##SNP Array Search 
- Currently lives at http://markermaker-148719.appspot.com
- used to search by: 
	- probeset ID
	- UCSC region
	- RS ID


### Starting development server
``` bash
~/bin/go_appengine/dev_appserver.py .
```

### Uploading to google app engine

```
 ~/bin/go_appengine/appcfg.py update .

```


### Populating data

```
curl -X POST -d "{\"markerName\":\"SNP_A-1863151\",\"rsID\":\"rs17054903\",\"chromosome\":\"3\",\"position\":55216428,\"a_allele\":\"A\",\"b_allele\":\"G\",\"arrays\":[\"Affymetrix SNP 6.0\"]}"  "http://markermaker-148719.appspot.com/populate/"
```



### Retrieving data

```
curl "http://markermaker-148719.appspot.com/probesetqRaw/SNP_A-1794291"
```
