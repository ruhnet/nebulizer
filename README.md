# Nebulizer

Nebulizer is a console application that takes JSON input and creates a set of certificate files for a [Nebula](https://github.com/slackhq/nebula) overlay network.

The ```nebula-cert``` binary is used to create the certificate files.

Input can be read from a JSON text file or piped in via standard input.

```
./nebulizer -f ./mynetwork.json
```

Example input:
```
{
  "ca": {
    "name":"My Nebula Overlay Network",
    "duration": 730
  },
  "hosts": [
    {
      "hostname": "lighthouse.nebula.mydomain.com",
      "ip": "172.31.9.1/26",
      "groups": []
    },
    {
      "hostname": "server1.nebula.mydomain.com",
      "ip": "172.31.9.2/26",
      "groups": [
        "servers",
        "app-backend"
      ]
    },
    {
      "hostname": "tmpadmin.nebula.mydomain.com",
      "ip": "172.31.9.8/26",
      "duration": 30,
      "groups": [
        "admin",
        "mod",
        "bobnet"
      ]
    },
    {
      "hostname": "laptop.nebula.mydomain.com",
      "ip": "172.31.9.5/26",
      "duration": 365,
      "groups": [
        "admin",
        "laptops",
        "mod",
        "bobnet"
      ]
    }
  ]
}
```

If you don't want to create the CA files, you can simply omit the ```ca``` object in the JSON.
The ```duration``` key is specified in days, and is always optional. If omitted for the CA, a default of 1 year is used. The default duration for hosts is until 1 second before the expiration of the CA.
Nebulizer will skip over creating a host or CA file if the certificate file already exists.

Run ```nebulizer -h``` to see the help:

```
Usage of ./nebulizer:
  -c string
    	CA certificate path. (default "./ca.crt")
  -f string
    	Path to network input file. Use '-' for standard input. (default "-")
  -k string
    	CA key path. (default "./ca.key")
  -o Overwrite existing files.
  -p string
    	Path to nebula-cert binary file. If not specified, search $PATH and current directory.
```


