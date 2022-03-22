<!--
SPDX-FileCopyrightText: 2022 2020-present Open Networking Foundation <info@opennetworking.org>

SPDX-License-Identifier: Apache-2.0
-->

**Table of Contents**
- [1. Introduction](#1-Introduction)
- [2. How to Install gNMI Command Line Interface (CLI) ?](#2-How-to-Install-gNMI-Command-Line-Interface-CLI)
- [3. Get the capabilities](#3-Get-the-capabilities)
- [4. Run the Get Command](#4-Run-the-Get-Command)
  - [4.1. Retrieve the motd-banner](#41-Retrieve-the-motd-banner)
  - [4.2. Retrieve All CONFIG leaves under "/system"](#42-Retrieve-All-CONFIG-leaves-under-%22system%22)
  - [4.3. Retrieve All STATE leaves under "/system"](#43-Retrieve-All-STATE-leaves-under-%22system%22)
  - [4.4. Retrieve All Config values under the root](#44-Retrieve-All-Config-values-under-the-root)
- [5. Run the Set command](#5-Run-the-Set-command)
- [6. Run the Subscribe command](#6-Run-the-Subscribe-command)
  - [6.1. Subscribe ONCE](#61-Subscribe-ONCE)
  - [6.2. Subscribe POLL](#62-Subscribe-POLL)
  - [6.3. Subscribe Stream](#63-Subscribe-Stream)
    - [6.3.1. ON\_CHANGE](#631-ONCHANGE)
    - [6.3.2. SAMPLE](#632-SAMPLE)
    - [6.3.3. TARGET\_DEFINED](#633-TARGETDEFINED)
  - [6.4. Generate and Stream Random Events for State Type Attributes (Just for **Testing** Purposes)](#64-Generate-and-Stream-Random-Events-for-State-Type-Attributes-Just-for-Testing-Purposes)
- [7. Troubleshooting](#7-Troubleshooting)
  - [7.1. Deadline exceeded](#71-Deadline-exceeded)
  - [7.2. TCP diagnosis](#72-TCP-diagnosis)
  - [7.3. HTTP Diagnosis](#73-HTTP-Diagnosis)
# 1. Introduction


# 2. How to Install gNMI Command Line Interface (CLI) ? 

gNMI CLI is a general purpose client tool for testing gNMI devices, from
the OpenConfig project.
To run it, two options are available:

1. (**Recommended Option**): you can install the gNMI CLI on your own machine using the following command and run it as an external application to the Docker containers. This option allows you to connect to any of the targets and run the gNMI CLI commands. 
```bash
go get -u github.com/openconfig/gnmi/cmd/gnmi_cli
go install -v github.com/openconfig/gnmi/cmd/gnmi_cli
```
2. Or you can ssh into any of the targets using the following command and run 
the gNMI CLI from the Docker container. 
```bash
docker exec -it <Container ID> /bin/bash
```

# 3. Get the capabilities
```bash
gnmi_cli -address localhost:10161 \
       -capabilities \
       -timeout 5s -alsologtostderr \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```

If you get
```bash
E0416 15:23:08.099600   22997 gnmi_cli.go:180] could not create a gNMI client: Dialer(localhost:10161, 5s): context deadline exceeded
```
It indicates a transport problem - see the [troubleshooting](#deadline-exceeded) section below.

# 4. Run the Get Command
## 4.1. Retrieve the motd-banner
The following command retrieves the motd-banner.
```bash
gnmi_cli -address localhost:10162 \
       -get \
       -proto "path: <elem: <name: 'system'> elem:<name:'config'> elem: <name: 'motd-banner'>>" \
       -timeout 5s -alsologtostderr \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```

This gives a response like
```bash
notification: <
  timestamp: 1555495881239352362
  update: <
    path: <
      elem: <
        name: "system"
      >
      elem: <
        name: "config"
      >
      elem: <
        name: "motd-banner"
      >
    >
    val: <
      string_val: "Welcome to gNMI service on localhost:10162"
    >
  >
>
```
## 4.2. Retrieve All CONFIG leaves under "/system"
```bash
gnmi_cli -address localhost:10162 \
       -get \
       -proto  "type:1, path: <elem: <name: 'system'>>" \
       -timeout 5s -alsologtostderr \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```

This gives a response like this:
```bash
notification: <
  timestamp: 1561659731806153000
  update: <
    path: <
      elem: <
        name: "system"
      >
    >
    val: <
      json_val: "{\"aaa\":{\"authentication\":{\"admin-user\":{\"config\":{\"admin-password\":\"password\"}},\"config\":{\"authentication-method\":[\"LOCAL\"]}}},\"clock\":{\"config\":{\"timezone-name\":\"Europe/Dublin\"}},\"config\":{\"domain-name\":\"opennetworking.org\",\"hostname\":\"replace-device-name\",\"login-banner\":\"This device is for authorized use only\",\"motd-banner\":\"replace-motd-banner\"},\"openflow\":{\"agent\":{\"config\":{\"backoff-interval\":5,\"datapath-id\":\"00:16:3e:00:00:00:00:00\",\"failure-mode\":\"SECURE\",\"inactivity-probe\":10,\"max-backoff\":10}},\"controllers\":{\"controller\":{\"main\":{\"config\":{\"name\":\"main\"},\"connections\":{\"connection\":{\"0\":{\"aux-id\":0,\"config\":{\"address\":\"192.0.2.10\",\"aux-id\":0,\"port\":6633,\"priority\":1,\"source-interface\":\"admin\",\"transport\":\"TLS\"}},\"1\":{\"aux-id\":1,\"config\":{\"address\":\"192.0.2.11\",\"aux-id\":1,\"port\":6653,\"priority\":2,\"source-interface\":\"admin\",\"transport\":\"TLS\"}}}},\"name\":\"main\"},\"second\":{\"config\":{\"name\":\"second\"},\"connections\":{\"connection\":{\"0\":{\"aux-id\":0,\"config\":{\"address\":\"192.0.3.10\",\"aux-id\":0,\"port\":6633,\"priority\":1,\"source-interface\":\"admin\",\"transport\":\"TLS\"}},\"1\":{\"aux-id\":1,\"config\":{\"address\":\"192.0.3.11\",\"aux-id\":1,\"port\":6653,\"priority\":2,\"source-interface\":\"admin\",\"transport\":\"TLS\"}}}},\"name\":\"second\"}}}}}"
    >
  >
>
```

## 4.3. Retrieve All STATE leaves under "/system"

```bash
gnmi_cli -address localhost:10162 \
       -get \
       -proto  "type:2, path: <elem: <name: 'system'>>" \
       -timeout 5s -alsologtostderr \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```

```bash
notification: <
  timestamp: 1561659901045020000
  update: <
    path: <
      elem: <
        name: "system"
      >
    >
    val: <
      json_val: "{\"openflow\":{\"controllers\":{\"controller\":{\"main\":{\"connections\":{\"connection\":{\"0\":{\"aux-id\":0,\"state\":{\"address\":\"192.0.2.10\",\"aux-id\":0,\"port\":6633,\"priority\":1,\"source-interface\":\"admin\",\"transport\":\"TLS\"}},\"1\":{\"aux-id\":1,\"state\":{\"address\":\"192.0.2.11\",\"aux-id\":1,\"port\":6653,\"priority\":2,\"source-interface\":\"admin\",\"transport\":\"TLS\"}}}},\"name\":\"main\"},\"second\":{\"connections\":{\"connection\":{\"0\":{\"aux-id\":0,\"state\":{\"address\":\"192.0.3.10\",\"aux-id\":0,\"port\":6633,\"priority\":1,\"source-interface\":\"admin\",\"transport\":\"TLS\"}},\"1\":{\"aux-id\":1,\"state\":{\"address\":\"192.0.3.11\",\"aux-id\":1,\"port\":6653,\"priority\":2,\"source-interface\":\"admin\",\"transport\":\"TLS\"}}}},\"name\":\"second\"}}}}}"
    >
  >
>
```

## 4.4. Retrieve All Config values under the root
For this case, we assume that when that path is empty but the dataType is 
specefied in the request, we return whole config data tree. 

```bash
gnmi_cli -address localhost:10162 \
       -get \
       -proto  "type:1" \
       -timeout 5s -alsologtostderr \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```

This gives a response like this:
```bash
notification: <
  timestamp: 1561660173314942000
  update: <
    path: <
    >
    val: <
      json_val: "{\"openconfig-interfaces:interfaces\":{\"interface\":[{\"config\":{\"name\":\"admin\"},\"name\":\"admin\"}]},\"openconfig-system:system\":{\"aaa\":{\"authentication\":{\"admin-user\":{\"config\":{\"admin-password\":\"password\"}},\"config\":{\"authentication-method\":[\"openconfig-aaa-types:LOCAL\"]}}},\"clock\":{\"config\":{\"timezone-name\":\"Europe/Dublin\"}},\"config\":{\"domain-name\":\"opennetworking.org\",\"hostname\":\"replace-device-name\",\"login-banner\":\"This device is for authorized use only\",\"motd-banner\":\"replace-motd-banner\"},\"openconfig-openflow:openflow\":{\"agent\":{\"config\":{\"backoff-interval\":5,\"datapath-id\":\"00:16:3e:00:00:00:00:00\",\"failure-mode\":\"SECURE\",\"inactivity-probe\":10,\"max-backoff\":10}},\"controllers\":{\"controller\":[{\"config\":{\"name\":\"main\"},\"connections\":{\"connection\":[{\"aux-id\":0,\"config\":{\"address\":\"192.0.2.10\",\"aux-id\":0,\"port\":6633,\"priority\":1,\"source-interface\":\"admin\",\"transport\":\"TLS\"}},{\"aux-id\":1,\"config\":{\"address\":\"192.0.2.11\",\"aux-id\":1,\"port\":6653,\"priority\":2,\"source-interface\":\"admin\",\"transport\":\"TLS\"}}]},\"name\":\"main\"},{\"config\":{\"name\":\"second\"},\"connections\":{\"connection\":[{\"aux-id\":0,\"config\":{\"address\":\"192.0.3.10\",\"aux-id\":0,\"port\":6633,\"priority\":1,\"source-interface\":\"admin\",\"transport\":\"TLS\"}},{\"aux-id\":1,\"config\":{\"address\":\"192.0.3.11\",\"aux-id\":1,\"port\":6653,\"priority\":2,\"source-interface\":\"admin\",\"transport\":\"TLS\"}}]},\"name\":\"second\"}]}}}}"
    >
  >
>
```

# 5. Run the Set command
The following command updates the timezone-name.  
```bash
gnmi_cli -address localhost:10161  \
       -set \
       -proto "update:<path: <elem: <name: 'system'>  elem: <name: 'clock' > elem: <name: 'config'> elem: <name: 'timezone-name'>> val: <string_val: 'Europe/Paris'>>"  \
       -timeout 5s \
       -alsologtostderr  \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```

This gives a response like this:
```bash
response: <
  path: <
    elem: <
      name: "system"
    >
    elem: <
      name: "clock"
    >
    elem: <
      name: "config"
    >
    elem: <
      name: "timezone-name"
    >
  >
  op: UPDATE
>
```

# 6. Run the Subscribe command
## 6.1. Subscribe ONCE
```bash
gnmi_cli -address localhost:10161 \
       -proto "subscribe:<mode: 1, prefix:<>, subscription:<path: <elem: <name: 'openconfig-system:system'>  elem: <name: 'clock' > elem: <name: 'config'> elem: <name: 'timezone-name'>>>>" \
       -timeout 5s -alsologtostderr \
       -client_crt certs/client1.crt -client_key certs/client1.key -ca_crt certs/onfca.crt
```

This gives a response like this. 
```bash
{
  "system": {
    "clock": {
      "config": {
        "timezone-name": "Europe/Dublin"
      }
    }
  }
}
```
## 6.2. Subscribe POLL
```bash
gnmi_cli -address localhost:10161 \
    -proto "subscribe:<mode: 2, prefix:<>, subscription:<path: <elem: <name: 'openconfig-system:system'>  elem: <name: 'clock' > elem: <name: 'config'> elem: <name: 'timezone-name'>>>>" \
    -timeout 5s -alsologtostderr \
    -polling_interval 5s \
    -client_crt certs/client1.crt -client_key certs/client1.key -ca_crt certs/onfca.crt
```
After running the above command the following output will be printed on the screen every 5 seconds. 
```bash
{
  "system": {
    "clock": {
      "config": {
        "timezone-name": "Europe/Dublin"
      }
    }
  }
}
```

## 6.3. Subscribe Stream
Stream subscriptions are long-lived subscriptions which continue to transmit updates relating to the set of paths that are covered within the subscription indefinitely. The current implementaiton of the simulator supports the following stream modes: 

### 6.3.1. ON\_CHANGE
When a subscription is defined to be "on change", data updates are only sent when the value of the data item changes. To test this mode, you should follow the following steps: 

1. First you need to run the following command to subcribe for the events on a path:
```bash
gnmi_cli -address localhost:10161 \
    -proto "subscribe:<mode: 0, prefix:<>, subscription:<mode:0, path: <elem: <name: 'openconfig-system:system'>  elem: <name: 'clock' > elem: <name: 'config'> elem: <name: 'timezone-name'>>>>" \
    -timeout 5s -alsologtostderr \
    -polling_interval 5s \
    -client_crt certs/client1.crt -client_key certs/client1.key -ca_crt certs/onfca.crt
```

After running the above command, you need to make a change in the timezone-name using set command as follows to get an update from the target about that change. 

```bash
gnmi_cli -address localhost:10161  \
       -set \
       -proto "update:<path: <elem: <name: 'system'>  elem: <name: 'clock' > elem: <name: 'config'> elem: <name: 'timezone-name'>> val: <string_val: 'Europe/Spain'>>"  \
       -timeout 5s \
       -alsologtostderr  \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```

The output in the terminal which runs subscribe stream will be like this: 
```bash
{
  "system": {
    "clock": {
      "config": {
        "timezone-name": "Europe/Spain"
      }
    }
  }
}
{
  "system": {
    "clock": {
      "config": {
        "timezone-name": "Europe/Spain"
      }
    }
  }
}
```

### 6.3.2. SAMPLE
A subscription that is defined to be sampled MUST be specified along with a *sample_interval* encoded as an unsigned 64-bit integer representing nanoseconds between samples. The target sends The value of the data item(s) once per sample interval to the client. For example, we would like to subscribe to receive *timezone-name* value from the gnmi target every 5 seconds. To do that, we can use the following command:
```bash
gnmi_cli -address localhost:10161  \
       "subscribe:<mode: 0, prefix:<>, subscription:<mode:2, sample_interval:5000000000 path: <elem: <name: 'system'>  elem: <name: 'clock' > elem: <name: 'config'> elem: <name: 'timezone-name'>>>>" \
       -timeout 5s \
       -alsologtostderr  \
       -client_crt certs/client1.crt \
       -client_key certs/client1.key \
       -ca_crt certs/onfca.crt
```
Every 5 seconds, the following output will be printed on the screen:
```bash
{
  "system": {
    "clock": {
      "config": {
        "timezone-name": "Europe/Dublin"
      }
    }
  }
}
```
The following assumptions have been made based on gNMI specefication to implement
the subscribe SAMPLE mode:

1. If the client sets the *sample_interval* to 0, the target uses the 
lowest sample interval (i.e. *lowestSampleInterval* variable) which is defined in target and has the default value of 5 seconds (i.e. 5000000000 nanoseconds). 

2. If the client sets the *sample_interval* to a value lower than *lowestSampleInterval* then the target rejects the request and returns an *InvalidArgument (3)* error code.

### 6.3.3. TARGET\_DEFINED
In the current version of the gnmi simulator, we define TARGET_DEFINED mode to behave 
always like ON_CHANGE mode. Accroding to the gNMI spec, the target MUST determine the best type of subscription to be created on a per-leaf basis. 


# 7. Troubleshooting

## 7.1. Deadline exceeded
If you get an error like
```bash
E0416 15:23:08.099600   22997 gnmi_cli.go:180] could not create a gNMI client:
Dialer(localhost:10161, 5s): context deadline exceeded
```

or anything about __deadline exceeded__, then it is **always** related to the
transport mechanism above gNMI i.e. TCP or HTTPS

## 7.2. TCP diagnosis
> This is not a concern with port mapping method using localhost and is for 
> the Linux specific option only

Starting with TCP - see if you can ping the device
1. by IP address e.g. 17.18.0.2 - if not it might not be up or there's some
   other network problem
2. by short name e.g. device1 - if not maybe your /etc/hosts file is wrong or
   DNS domain search is not opennetworking.org
3. by long name e.g. device1.opennetworking.org - if not maybe your /etc/hosts
   file is wrong

For the last 2 cases make sure that the IP address that is resolved matches what
was given at the startup of the simulator with docker.

## 7.3. HTTP Diagnosis
If TCP shows reachability then try with HTTPS - it's very important to remember
that for HTTPS the address at which you access the server **must** match exactly
the server name in the server key's Common Name (CN) like __localhost__ or
__device1.opennetworking.org__ (and not an IP address!)

Try using cURL to determine if there is a certificate problem
```
curl -v https://localhost:10164 --key certs/client1.key --cert certs/client1.crt --cacert certs/onfca.crt
```
This might give an error like
```bash
* Rebuilt URL to: https://localhost:10163/
*   Trying 172.18.0.3...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 10163 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: certs/onfca.crt
  CApath: /etc/ssl/certs
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Request CERT (13):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Certificate (11):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS handshake, CERT verify (15):
* TLSv1.2 (OUT), TLS change cipher, Client hello (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-AES256-GCM-SHA384
* ALPN, server accepted to use h2
* Server certificate:
*  subject: C=US; ST=CA; L=MenloPark; O=ONF; OU=Engineering; CN=device3.opennetworking.org
*  start date: Apr 16 14:40:46 2019 GMT
*  expire date: Apr 15 14:40:46 2020 GMT
* SSL: certificate subject name 'device3.opennetworking.org' does not match target host name 'localhost'
* stopped the pause stream!
* Closing connection 0
* TLSv1.2 (OUT), TLS alert, Client hello (1):
curl: (51) SSL: certificate subject name 'device3.opennetworking.org' does not match target host name 'localhost'
```

> In this case the device at __localhost__ has a certificate for
> device3.opennetworking.org. HTTPS does not accept this as a valid certificate
> as it indicates someone might be spoofing the server. This happens today in
> your browser if you access a site through HTTPS whose certificate CN does not
> match the URL - it is just a fact of life with HTTPS, and is not peculiar to gNMI.

Alternatively a message like the following can occur:
```bash
* Rebuilt URL to: https://onos-config:5150/
*   Trying 172.17.0.4...
* TCP_NODELAY set
* Connected to onos-config (172.17.0.4) port 5150 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: ../simulators/pkg/certs/onfca.crt
  CApath: /etc/ssl/certs
* (304) (OUT), TLS handshake, Client hello (1):
* (304) (IN), TLS alert, Server hello (2):
* error:14094410:SSL routines:ssl3_read_bytes:sslv3 alert handshake failure
* stopped the pause stream!
* Closing connection 0
curl: (35) error:14094410:SSL routines:ssl3_read_bytes:sslv3 alert handshake failure
```
> This could mean many things - e.g. that the cert on the server is empty or
> that the Full Qualified Domainname (FQDN) of the device does not match the
> subject of the certificate.

When device names and certificates match, then curl will reply with a message like:
```bash
curl: (92) HTTP/2 stream 1 was not closed cleanly: INTERNAL_ERROR (err 2)
```

> This means the HTTPS handshake was __successful__, and it has failed at the
> gNMI level - not surprising since we did not send it any gNMI payload. At this
> stage you should be able to use **gnmi_cli** directly.
