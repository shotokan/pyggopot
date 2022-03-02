# Pyggpot Microservice Sample

In case you're wondering: [Pyggpot](https://bahoukas.com/pygg-pots-to-piggy-banks/)

## Prereqs

- [Go 1.16](https://golang.org/doc/install). Go < 1.16 _probably_ won't work due to [this issue in xoutil](https://github.com/xo/xo/issues/242).

## Installation

- Clone repo

```$bash
git clone git@github.com:aspiration-labs/pyggpot.git
```

- Install modules and vendored tools

```$bash
make setup
```

- Create sqlite database

```$bash
make db
```

- Generate proto and model code

```$bash
make all
```

## Run

```$bash
go run cmd/server/main.go
```

## Play

Swagger site at [http://localhost:8080/swaggerui/](http://localhost:8080/swaggerui/)

## TODO Document a better design to support the RemoveCoins service

The current design is not scalable since information is duplicated which means that more work has to be done when running RemoveCoins.

Currently we have a number of known or fixed denominations, however this should be flexible to change so we must have a scheme where the denominations and their value are stored, denomination schema.

The provider must be implemented in order to insert the information into the new naming scheme.

Another change that needs to be made is to have a registration per denomination tied to a pot. That is, in the coins table, a denomination must not be repeated several times (several records) for the same pot, so when we insert a coin for a pot, it must be verified if that denomination already exists for that pot it will be inserted otherwise the count for that denomination would be updated.

When the records with the denominations reach 0, that is, there are no more coins in a pot of one denomination, those records must be deleted.

You could have a query to fetch the currency records that are going to be used in an array so as not to consult them one by one, since due to the nature of the problem, a large amount of data will not be handled since the pots have a limit as well as the coins to be mined.

Transactions should be used to deal with concurrency.
