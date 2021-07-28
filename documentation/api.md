# API Documentation

The intention of this document is to describe how to make requests to the API and what to expect
from it.

### Table of Contents

* [About Permissions](#about-permissions)
* [User Based Actions](#user-based-actions)
    * [Login](#login)
    * [Change Password](#change-password)
* [User Referrals](#user-referrals)
    * [Creating a Referral Code](#creating-a-referral-code)
    * [Using a Referral Code](#using-a-referral-code)
    * [Listing Active Referral Codes](#listing-active-referral-codes)

## About Permissions

Before reading any further documentation it must be made clear that anything requiring permissions
for an authenticated user gets overridden in the following scenarios. Therefore, if documentation
below says you need a certain permission to take an action, relevant exceptions here apply or
override those below without being explicitly stated. The rules are sequential, meaning that each
rule takes precedence over the next.

1. The Admin user created by the framework itself is a permanent administrator in
   **every** scenario. They have truly complete and full power over everything. Nobody can revoke
   their permissions. **Be careful and don't let others have access to this!**

1. The owner of a server has full control of anything at that server level regardless of permissions
   assigned to them. They are allowed to take any actions they want on a server they own.

1. If the user has the user level permission `administrator`, they have full control of everything
   to the web server and all game servers on it. Any permission checks will automatically pass for
   them. In addition, this user is able to assign and remove administrator to other users as
   well. **Be careful who you trust with this permission!**

1. If the user has a server level permission `administrator`, they have full control of everything
   at the game server level. Similar to user level permission 'administrator', you can add and
   revoke said permission from every other user in the portal for this particular game server.

## User Based Actions

This section will describe actions outside the regular `/api` endpoint. Things such as login, and
changing your password will appear here

### Login

First and foremost if logging into the server. The actual API is unavailable to the user if they do
not log in.

#### Format

You must send an HTTP POST request to `/login` with the form data containing your username and
password. If the username and password pair are valid, the server will respond with a status found
and redirect you to the login page. It will also return a token in the form of a cookie to be used
for future authentication. If the login is invalid, you will get an HTTP Status Forbidden error

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

As a user who knows their password, you can change your password by hitting this endpoint.
Currently, there does not exist a way to recover a password that is lost so don't forget it!

#### Format

You must send an HTTP POST request to `/change-password` with the form data containing your current
password, and your new password. If the current password is invalid, you will get an HTTP Status
Forbidden error. If the request is valid, the server will respond with a status found and redirect
you to the index page.

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

This section will describe how to invite a user to the web portal. In order to get an account on
this framework, you must create a referral code to give to a user. This will allow them to use the
referral code to create an account themselves.

### Creating a Referral Code

To create a referral code, you must first be an authenticated user and have the
permission `invite_user`. Without this, you will receive either an HTTP Status Unauthorized or HTTP
Status Forbidden error. Referral codes are single use and expire after 24 hours.

#### Format

Given you have permissions, submitting a POST request to `/api/refer` with an empty body, and a
valid token cookie will return a JSON response with the code in it.

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

{"code":92659391,"status":"Success"}
```

### Using a Referral Code

To use a referral code, you must have a valid, unexpired referral code. It will be consumed upon
use. If you submit a request to an invalid code, you will receive an HTTP Status Bad Request Error

#### Format

Submit a POST request to `/api/refer/:code` where `:code` is your referral code. The body of this
POST should contain the fields `username` and `password`. Your `username` must not conflict with any
other existing usernames in the database.

#### Example

Request:

```http request
POST /api/refer/92659391 HTTP/1.1
Host: localhost:8080
Accept: application/json

{"username": "stew3254", "password": "UNuD3TUWgiBPU2!r2U6X"}
```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 20
Content-Type: application/json
Date: Wed, 28 Jul 2021 20:46:32 GMT

{"status":"Success"}
```

### Listing Active Referral Codes

There are two ways to list active referral codes. If you know the code, you can either get the
entire list of codes, or you can get information about a specific code

#### Format

Submit a GET request to `/api/refer/` to get a listing of all valid authentication codes. This
request requires the `invite_user` permission to do. The reason being, then an authenticated user
without that cannot just troll that endpoint waiting for codes to come up to invite other users by
stealing a code someone else made.

The second is to submit a GET request to `/api/refer/:code` if that code is valid, it will return
details about the code. This does **not** need the `invite_user` request to do. The reason being is
that if this permission is required, a malicious user without authentication could create
intentionally malformed POST requests to send to each code url. If they receive an HTTP Bad Request
Status, then they know the code exists. Therefore, hiding this response wouldn't help much anyways.

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
Content-Length: 179
Content-Type: application/json
Date: Wed, 28 Jul 2021 20:46:32 GMT

[{"code":42910757,"expiration":"2021-07-29T21:44:05.740188Z","user":{"username":"admin"}},{"code":99927865,"expiration":"2021-07-29T21:44:25.763361Z","user":{"username":"admin"}}]
```

Single Code Request:

```http request
GET /api/refer/37956766 HTTP/1.1
Host: localhost:8080
Cookie: token=8e293bfe1e482996f0782d4caac775d6cec81a102885547c939279f0e6634785
Accept: application/json

```

Response:

```http request
HTTP/1.1 200 OK
Content-Length: 20
Content-Type: application/json
Date: Wed, 28 Jul 2021 20:46:32 GMT

{"code":37956766,"expiration":"2021-07-29T20:23:45.98591Z","user":{"username":"admin"}}}
```
