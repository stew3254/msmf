# API Documentation

The intention of this document is to describe how to make requests to the API
and what to expect from it.

### Table of Contents

* [About Permissions](#about-permissions)
* [User Based Actions](#user-based-actions)
    * [Login](#login)
    * [Change Password](#change-password)
* [User Referrals](#user-referrals)
    * [Creating Referral Codes](#creating-referral-codes)
    * [Using Referral Codes](#using-referral-codes)
    * [Listing Active Referral Codes](#listing-active-referral-codes)
* [Server Management](#server-management)
    * [Creating Servers](#creating-servers)
    * [Viewing Servers](#viewing-servers)
    * [Updating Servers](#updating-servers)
    * [Starting and Stopping Servers](#starting-and-stopping-servers)
    * [Deleting a Server](#deleting-a-server)
    * [Connecting to Server Consoles](#connecting-to-server-consoles)
    * [Creating Discord Integrations](#creating-discord-integrations)
    * [Viewing Discord Integrations](#viewing-discord-integrations)
    * [Deleting Discord Integrations](#deleting-discord-integrations)

## About Permissions

Before reading any further documentation it must be made clear that anything
requiring permissions for an authenticated user gets overridden in the following
scenarios. Therefore, if documentation below says you need a certain permission
to take an action, relevant exceptions here apply or override those below
without being explicitly stated. The rules are sequential, meaning that each
rule takes precedence over the next.

1. The Admin user created by the framework itself is a permanent administrator
   in
   **every** scenario. They have truly complete and full power over everything.
   Nobody can revoke their permissions. **Be careful and don't let others have
   access to this!**

1. The owner of a server has full control of anything at that server level
   regardless of permissions assigned to them. They are allowed to take any
   actions they want on a server they own.

1. If the user has the user level permission `administrator`, they have full
   control of everything to the web server and all game servers on it. Any
   permission checks will automatically pass for them. In addition, this user is
   able to assign and remove administrator to other users as well. **Be careful
   who you trust with this permission!**

1. If the user has a server level permission `administrator`, they have full
   control of everything at the game server level. Similar to user level
   permission `administrator`, you can add and revoke said permission from every
   other user in the portal for this particular game server.

## User Based Actions

This section will describe actions outside the regular `/api` endpoint. Things
such as login, and changing your password will appear here

### Login

First and foremost is logging into the server. The actual API is unavailable to
the user if they do not log in.

#### Format

You must send an HTTP POST request to `/login` with the form data containing
your username and password. If the username and password pair are valid, the
server will respond with a status found and redirect you to the index page. It
will also return a token in the form of a cookie to be used for future
authentication. If the login is invalid, you will get an HTTP Status
Unauthorized

#### Example

Request:

```http request
POST /login HTTP/1.1
Host: localhost:8080
Content-Type: application/x-www-form-urlencoded

username: "user"
password: "password"
```

Response:

```http request
HTTP/1.1 302 Found
Content-Length: 0
Date: Wed, 28 Jul 2021 15:10:05 GMT
Location: /
Set-Cookie: token=33dffe8cb1743d9ddfe82bd2b3caeb3510c3020ddae2f6770cbfb26103be2c32; Expires=Wed, 28 Jul 2021 21:10:05 GMT; HttpOnly; Secure; SameSite=Strict
```

### Change Password

As a user who knows their password, you can change your password by hitting this
endpoint. Currently, there does not exist a way to recover a password that is
lost so don't forget it!

#### Format

You must send an HTTP POST request to `/change-password` with the form data
containing your current password, and your new password. If the current password
is invalid, you will get an HTTP Status Forbidden error. If the request is
valid, the server will respond with a status found and redirect you to the index
page.

#### Example

Request:

```http request
POST /change-password HTTP/1.1
Host: localhost:8080
Content-Type: application/x-www-form-urlencoded

current_password: "bad password"
new_password: "T0tally_$uper S3cur3 P@ssword!."
```

Response:

```http request
HTTP/1.1 302 Found
Content-Length: 0
Date: Wed, 28 Jul 2021 15:11:00 GMT
Location: /
```

## User Referrals

This section will describe how to invite a user to the web portal. In order to
get an account on this framework, you must create a referral code to give to a
user. This will allow them to use the referral code to create an account
themselves.

### Creating Referral Codes

To create a referral code, you must first be an authenticated user and have the
permission `invite_user`. Without this, you will receive either an HTTP Status
Unauthorized or HTTP Status Forbidden error. Referral codes are single use and
expire after 24 hours.

#### Format

Given you have permissions, submitting a POST request to `/api/refer` with an
empty body, and a valid token cookie will return a JSON response with the code
in it.

#### Example

Request:

```http request
POST /api/refer HTTP/1.1
Host: localhost:8080
Cookie: token=8e293bfe1e482996f0782d4caac775d6cec81a102885547c939279f0e6634785
Accept: application/json
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 36
Content-Type: application/json
Date: Wed, 28 Jul 2021 20:23:59 GMT

{
  "code": 39363939,
  "status": "Success"
}
```

### Using Referral Codes

To use a referral code, you must have a valid, unexpired referral code. It will
be consumed upon use. If you submit a request to an invalid code, you will
receive an HTTP Status Bad Request Error

#### Format

Submit a POST request to `/api/refer/:code` where `:code` is your referral code.
The body of this POST should contain the fields `username` and `password`.
Your `username` must not conflict with any other existing usernames in the
database.

#### Example

Request:

```http request
POST /api/refer/92659391 HTTP/1.1
Host: localhost:8080
Accept: application/json

{
  "username": "stew3254",
  "password": "UNuD3TUWgiBPU2!r2U6X"
}
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 20
Content-Type: application/json
Date: Wed, 28 Jul 2021 20:46:32 GMT

{
    "status": "Success"
}
```

### Listing Active Referral Codes

There are two ways to list active referral codes. If you know the code, you can
either get the entire list of codes, or you can get information about a specific
code

### Structure

Here is a top-down view of the Referral object:

* `username` - The name of the user who made the object
* `username` - The name of the user who made the object
* `username` - The name of the user who made the object

#### Format

Submit a GET request to `/api/refer/` to get a listing of all valid
authentication codes. This request requires the `invite_user` permission to do.
The reason being, then an authenticated user without that cannot just troll that
endpoint waiting for codes to come up to invite other users by stealing a code
someone else made.

The second is to submit a GET request to `/api/refer/:code` if that code is
valid, it will return details about the code. This does **not** need
the `invite_user` request to do. The reason being is that if this permission is
required, a malicious user without authentication could create intentionally
malformed POST requests to send to each code url. If they receive an HTTP Bad
Request Status, then they know the code exists. Therefore, hiding this response
wouldn't help much anyways.

#### Examples

All Codes Request:

```http request
GET /api/refer HTTP/1.1
Host: localhost:8080
Cookie: token=8e293bfe1e482996f0782d4caac775d6cec81a102885547c939279f0e6634785
Accept: application/json
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 267
Content-Type: application/json
Date: Sat, 31 Jul 2021 02:22:15 GMT

[
  {
    "code": 12199222,
    "expiration": "2021-08-01T02:18:46.919706Z",
    "user": {
      "username": "admin"
    }
  },
  {
    "code": 39363939,
    "expiration": "2021-08-01T02:18:50.693649Z",
    "user": {
      "username": "admin"
    }
  },
  {
    "code": 20769021,
    "expiration": "2021-08-01T02:21:27.26003Z",
    "user": {
      "username": "admin"
    }
  }
]
```

Single Code Request:

```http request
GET /api/refer/39363939 HTTP/1.1
Host: localhost:8080
Cookie: token=8e293bfe1e482996f0782d4caac775d6cec81a102885547c939279f0e6634785
Accept: application/json
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 88
Content-Type: application/json
Date: Sat, 31 Jul 2021 02:25:40 GMT

{
  "code": 39363939,
  "expiration": "2021-08-01T02:18:50.693649Z",
  "user": {
    "username": "admin"
  }
}
```

## Server Management

This section discusses all things to do with management of video game servers.
This will include websockets and Discord integration.

### Creating Servers

To create a server, you must be authenticated and have the valid user level
permission
`create_server`.

#### Format

Submit a JSON POST request to `/api/server` with the following required
parameters:

* `name` - This is the name of your server. You must not have this name for any
  other servers you own.
* `game` - This is the game your server is for. Currently, the only supported
  game is 'Minecraft'.
* `port` - This is the port your server is accessible on. These must not
  conflict with any other ports reserved to other servers.

The following parameters are optional, but can potentially change the way the
server is created:

* `version` - This is the version tag of the game you are playing. This is
  important if you play games like Minecraft where you may want mods on versions
  that are incompatible with the latest edition.

The post body may also contain optional parameters which will be passed in as
environmental variables at the time of creation. In the case of Minecraft for
example, if you passed in the key `memory` with the value `2G`, that would add
the environmental variable `MEMORY=2G` to the container to give your server more
RAM. To know whether you should add environmental variables or not, look up the
documentation of the container being used to host the server per game.

#### Example

Request:

```http request
POST http://localhost:8080/api/server
Content-Type: application/json
Cookie: token=0f83e3dca3d29fdfdda77054061ba787c56b1235448e988cf483ca079ccd2df7

{
  "name": "Vanilla",
  "port": 25565,
  "game": "Minecraft",
  "version": "1.16.5",
  "motd": "Something is here"
}
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 20
Content-Type: application/json
Date: Thu, 29 Jul 2021 01:18:43 GMT

{
  "status": "Success"
}
```

### Viewing Servers

To view a server, you must meet any of the following conditions:

* You are the owner of a server
* You have any server level permission for that server
* You have any of the following user level permissions:
    * `administrator`
    * `manage_server_permission`
    * `delete_server`

If you do not meet any of these requirements, you will get an HTTP Status Not
Found error when trying to look at a server directly, and this server will not
show up in the all servers listing.

### Structure

Here is a top-down view of the server object.

* `id` - The unique identifier of the server. You'll see this number in the url.
* `port` - The port the server can be accessed on. This is unique.
* `name` - The name of the server. There can be more than one server named the
  same thing, but each user can only use a specific name once.
* `running` - Whether the server is currently running or not.
* `game` - The game the server is made for.
    * `name` - The name of the game the server is.
* `owner` - The owner of the server.
    * `name` - The name of the owner of the server.
* `version` - The version of the server if it is relevant
    * `tag` - The version tag

#### Format

To view all servers, submit a GET request to `/api/server`. You can specify the
column you would like to order on by using the `order_by` query string
parameter. If you would then like to reverse the ordering, set
`reverse=true` in the query string

If you would like to view a single server, make a GET request to
`/api/server/:server_id`

#### Examples

```http request
GET /api/server HTTP/1.1
Host: localhost:8080
Cookie: token=6bbae15cc44adf688331c21c66670dfd40469dedf9a29e8e45ad0440dbb6db2a
Accept: application/json
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 139
Content-Type: application/json
Date: Fri, 30 Jul 2021 20:46:36 GMT

[
  {
    "id": 1,
    "port": 25565,
    "name": "Vanilla",
    "running": false,
    "game": {
      "name": ""
    },
    "owner": {
      "username": ""
    },
    "version": {
      "tag": "",
      "game": {
        "name": ""
      }
    }
  }
]
```

Viewing a single server:

```http request
GET /api/server/1 HTTP/1.1
Host: localhost:8080
Cookie: token=6bbae15cc44adf688331c21c66670dfd40469dedf9a29e8e45ad0440dbb6db2a
Accept: application/json
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 157
Content-Type: application/json
Date: Fri, 30 Jul 2021 20:48:28 GMT

{
  "id": 1,
  "port": 25565,
  "name": "Vanilla",
  "running": false,
  "game": {
    "name": "Minecraft"
  },
  "owner": {
    "username": "admin"
  },
  "version": {
    "tag": "1.16.5",
    "game": {
      "name": ""
    }
  }
}
```

### Updating Servers

In order to update a server, you must have meet any of the following conditions:

* You are owner of the server
* You have the server level permissions `administrator` or `edit_configuration`
* You have the user level permissions `administrator` or
  `manage_server_permission`

#### Format

Submit a JSON PATCH request to `/api/server/:server_id`. You may set any of the
following parameters:

* `port` - You cannot change the port to one that has already been allocated
* `name` - You cannot change the name to the name of a server already owned
* by the owner
* `version_tag` - This can be freely set

#### Example

Request:

```http request
PATCH http://localhost:8080/api/server/1
Content-Type: application/json
Cookie: token=3e93edc18cd45111bf06389f2f1b1e00976a2d4818932fec34e79ba576d189bb

{
  "name": "Vanilla",
  "port": 25565
}
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 157
Content-Type: application/json
Date: Sat, 31 Jul 2021 03:13:35 GMT

{
  "id": 1,
  "port": 25565,
  "name": "Vanilla",
  "running": false,
  "game": {
    "name": "Minecraft"
  },
  "owner": {
    "username": "admin"
  },
  "version": {
    "tag": "1.16.5",
    "game": {
      "name": ""
    }
  }
}
```

### Starting and Stopping Servers

In order to start, stop or restart a server, you must have meet any of the
following conditions:

* You are owner of the server
* You have the server level permissions `administrator` or `restart`
* You have the user level permissions `administrator` or
  `manage_server_permission`

#### Format

Create an empty POST request to any of the following to get your desired effect:

* `/api/server/:server_id/start`
* `/api/server/:server_id/stop`
* `/api/server/:server_id/restart`

As long as the server actually exists, you should get no errors. It will return
a successful response once the intended action happens. However, if you try to
chain too many of these in a row, your requests might time out before they get
fulfilled.

### Deleting a Server

In order to delete a server, you must have meet any of the following conditions:

* You are owner of the server
* You have the user level permissions `administrator` or `delete_server`

#### Format

Submit a DELETE request to `/api/server/:server_id`. It will either succeed or
you don't have valid permissions to do this action

### Connecting to Server Consoles

Server consoles allow you to directly interact with the server and send commands
to it. It often also gives you a way to chat with your players as they are
playing. Anyone with access to this can become a server administrator, so it
should only be given to people you trust. In order to view a server console, you
must have any of the following permissions:

* You are owner of the server
* You have the server level permissions `administrator` or
* `manage_server_console`
* You have the user level permissions `administrator` or
  `manage_server_permission`

#### Format

Create a websocket connection to `/api/ws/server/:server_id`.

In order to send a message to stdin of the server, send the message with type
`Text` to the server. This message will be sent back to you by default and every
other existing websocket connection to this server before being sent to stdin.
This is to ensure other clients with the server console open understand that
there was a message sent to get why a potentially confusing output appeared on
their screen. If you do not want to see your message played back to you like in
the case of terminal based viewers, send a message of type `Ping` with the
contents `no-repeat` verbatim. This will stop you from seeing your own messages.
To turn the repeat feature back on, sent a message of type `Ping` with the
contents `repeat` verbatim.

When reading from the websocket, you will receive a message of type `Text`
or of type `Binary`. Messages of type `Text` come from stdout, while messages of
type `Binary` come from stderr. Every open websocket connection to this server
will receive these messages in identical order while their connection is open at
the same time.

On connect, the web server will send up to the last 100 lines of history
available since its start. This will all appear before any live data is sent to
the socket

In order to keep the connection alive over an extended period of time, the
client should implement a function to ping the server every 5-10 seconds. Said
function should send a message of type `Ping` to the server with any contents
other than `no-repeat` or `repeat`. Upon receiving a successful ping, the server
will reply `Pong!` with a message of type `Pong`.

#### Example

TODO: Someone please add examples. HTTP example seems weird, so do Javascript or
something. A Go one might be added eventually

### Creating Discord Integrations

TODO Finish this

#### Format

#### Example

### Viewing Discord Integrations

TODO Finish this

#### Format

#### Example

### Deleting Discord Integrations

TODO Finish this

#### Format

#### Example
