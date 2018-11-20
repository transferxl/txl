# txl - command line for TransferXL

Accelerated transfers from the command line for TransferXL. Transfers of up to 1 TB in size are supported.

## Usage 

```
$ txl put -u user@gmail.com -p REDACTED videos.zip
https://transferxl.com/00fdva6dgdsf3rsw
$ txl get https://transferxl.com/00fdva6dgdsf3rsw
Downloaded 'videos.zip' in 54.6s (30.9 GB at 4.54 Gbit/s)
```

## Performance

On an EC2 instance with 10 Gbit networking capability, the following transfer speeds have been measured (us-east region).

| c4.8xlarge (10 Gbit) | Round trip |  Upload | Download | Unit     |
|:---------------------| ----------:| -------:| --------:|:---------| 
| 100 GB               |    14:15.8 | 11:26.1 |  02:49.7 | min:s    |
|                      |     *1.87* |  *1.17* |   *4.71* | *Gbit/s* |
| 200 GB               |    29:01.3 | 22:27.6 |  06:33.7 | min:s    |
|                      |     *1.84* |  *1.19* |   *4.06* | *Gbit/s* |
| 300 GB               |    42:24.7 | 32:52.8 |  09:31.9 | min:s    |
|                      |     *1.89* |  *1.22* |   *4.20* | Gbit/s   |

See `performance.sh` in order to do measure your own speeds.

## Download or build from source

Download a prebuilt [binary](https://github.com/transferxl/txl/releases/tag/v0.9-prerelease) or build your own: make sure Golang is [installed](https://golang.org/dl/) and do as follows:

```
$ go get -u github.com/transferxl/txl
```

## Encrypted transfers

Transfers can be protected using encryption by supplying a password phrase while creating the transfer.
In order to successfully download the transfer again it is necessary to pass in the password.
Encryption is possible for all transfers irrespective of the content of the transfer. 

```
$ txl put -e "right" secret.zip 
https://transferxl.com/00agdsf6dfdv3rsw
$ txl get https://transferxl.com/00agdsf6dfdv3rsw
Bad Request.
$ txl get -d "wrong" https://transferxl.com/00agdsf6dfdv3rsw
Access Denied.
$ txl get -d "right" https://transferxl.com/00agdsf6dfdv3rsw
Downloaded 'secret.zip' in 4.9s (1.9 GB at 312 Mbit/s)
```

Note that encryption does not store the password somewhere on the server or in the cloud (feel free to check the source code).
Therefore it is the responsibility of the sender to keep track of the password and pass it along to any downloaders.
A lost password means that you are no longer able to download a transfer.

In the event that the url of the transfer falls into the wrong hands, no downloads are possible unless the attacker also gets a hold of the decryption password.

## Subscription

In order to be able to send transfers, you need to have a subscription (Enterprise or up), see [plans](https://transferxl.com/plans).

## Other examples

```
$ zip -r - *.jpg | txl put -m "Here are the photos"
```

## Full command line syntax

```
$ txl help
Command line interface for TransferXL.com

Usage:
  txl [flags]
  txl [command]

Available Commands:
  put         Upload transfer
  get         Download transfer
  list        List transfers
 
Flags:
  -h, --help   help for txl

Use "txl [command] --help" for more information about a command.
```

```
$ txl put --help
Create a transfer

Usage:
  txl put  [flags]

Flags:
  -e, --encrypt string      encryption phrase
  -h, --help                help for put
  -l, --log string          file for logging
  -m, --message string      message for the transfer
  -p, --password string     password for account
  -r, --recipients string   email address of recipient(s)
  -s, --storage string      storage region for the transfer
  -u, --user string         user account
  -v, --verbose             verbose output
```

```
$ txl get --help                                                                                                                                                         
Get a transfer

Usage:
  txl get [flags]

Flags:
  -d, --decrypt string   decryption phrase
  -h, --help             help for get
  -l, --log string       file for logging
  -o, --output string    file for output
  -v, --verbose          verbose output
```

```
$ txl list --help
List all transfers

Usage:
  txl list [flags]

Flags:
  -h, --help              help for list
  -p, --password string   password for account
  -u, --user string       user account
```

## License

This code is licensed under the Apache License 2.0.
