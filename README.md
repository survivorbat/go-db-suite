# ⛁ DB Suites

This library contains a embeddable DB suite that will use [testcontainers](https://golang.testcontainers.org/modules/postgres/) to spin up a database server and
create a fresh isolated database per test.

## ⬇️ Installation

`go get github.com/survivorbat/go-db-suite`

## 📋 Usage

```go
import (
  "testing"

  "github.com/stretchr/testify/suite"
  "github.com/survivorbat/go-db-suite"
)

type CupcakeRepositorySuite struct {
  dbsuite.Postgres
}

func (c *CupcakeRepositorySuite) TestGetAll() {
  // Arrange
  repository := &CupcakeRepository{Db: c.DB}
  // Act
  data, err := repository.GetAll(t.Context())
  // Assert
  // [...]
}

func TestCupcakeRepositorySuite(t *testing.T) {
  t.Parallel()
  suite.Run(t, new(CupcakeRepositorySuite))
}
```

To add additional setup code, such as running migrations:

```go
import (
  "testing"

  "github.com/stretchr/testify/suite"
  "github.com/survivorbat/go-db-suite"
)

type CupcakeRepositorySuite struct {
  dbsuite.Postgres
}

func (c *CupcakeRepositorySuite) SetupTest() {
  // Important to call the embedded setup method
  c.Postgres.SetupTest()

  // Run migrations here
}

func TestCupcakeRepositorySuite(t *testing.T) {
  t.Parallel()
  suite.Run(t, new(CupcakeRepositorySuite))
}
```

## 🔭 Plans

- More configuration options
- Examples on how to use it with subtests
- More databases
