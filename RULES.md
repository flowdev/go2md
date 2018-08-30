# The FlowDev Rules For the Go Programming Language

The aim of these rules is to provide Go programmers with a simple yet powerful set of rules that
result in modular and idiomatic code when followed.
This should especially hold true in big and complex projects.

## Rules For Stateless Components
Stateless components can be simply implemented with free functions.
If there is only one input port (`in`) it is idiomatic to omit its name from the function and simply name the function like the component.

### Only One Output Port
If there is only one output port (`out`) it is idiomatic to omit its name from the return values.
```go
func addPersonalData(p *Person, name string, age int) *Person
```
This means that there are already many valid components out there and many libraries can be used without any wrappers.

### Special Case Error Output Port
If the last return value is of type `error` it is implicitly an extra port `err`.
It is valid but not necessary to name this value `err` or `portErr`.
The value for this port has to be handled first and the other values are only valid if the `error` value is `nil`.
```go
func addCustomerFromDB(o *Order) (*Order, error)
```
This again means that there are already many valid components out there and many libraries can be used without any wrappers.
In the Go standard library there are a few places where valid data is returned additionally to an error (i.e.: `io.Reader`).
These places are exceptions and this rule is very much in line with rules for idiomatic Go code.

### Multiple Output Ports
In case of multiple output ports we need to be able to distinguish them easily.
This is achived with a simple naming convention:

Each return value has a name that starts with `port` and is followd with an upper case letter and optionally more characters.
```go
func protectFromFraught(o *Order) (portOut *Order, portFraught *Order, err error)
```
The only hard requirement here is that the return value for each must be `nil` if the port isn't used.
The `error` type is again fulfilling this.

Multiple return values per port can be wrapped in a struct:
```go
func protectFromFraught(
	o *Order,
	c *Customer,
) (portOut *Order, portFraught *struct{o *Order, c *Customer}, portErr error)
```
Components with multiple output ports don't look like idiomatic Go code and can't be found like this in libraries on GitHub.
Fortunately such components are quite rare. But they are very useful when needed.

### Multiple Input Ports
In case of multiple input ports we need to be able to distinguish them easily.
This is achived with a simple naming convention:

Each function has a name that consists of the name of the component followed by the word `Port` and the name of the port.
```go
func addPersonalDataPortIn(p *Person, name string, age int) *Person
func addPersonalDataPortAddress(p *Person, a *Adress) *Person
```
Components with multiple input ports don't look like idiomatic Go code and can't be found like this in libraries on GitHub.
But it is quite easy and intuitive to understand once we know about components, ports and flows at all.
Again such components are quite rare but very useful when needed.

## Rules For Stateful Components
If we need to keep state between instances of handling data of input ports, well we need a data type to hold it.
This is exactly idiomatic Go. There are no special rules about the data type to use.
The data type that holds the state should be named like the component as a noun.

For the methods that can be called on the data type exactly the same rules apply as for the stateless components.
```go
type addCustomerFromDBer sql.Conn

func (c addCustomerFromDBer) addCustomerFromDB(o *Order) (*Order, error)
```

## Plugins As a Compromise
Sometimes we want to split up a complex component or pull out parts of it for better testability.
Turning all parts into components and making all the connections explicit can become unwieldy because it would distract too much from the big picture.
The general concept of how the flow works would get lost because of too many less important connections.

For situations like these plugins can be a nice solution.
Plugins are normal components. So they can be stateless or stateful as necessary.
But they can be handed over to input ports of other components after the normal parameters.
```go
func enrichOrderWithCustomerData(
	o *Order,
	pluginGetCustomer func(id string)*Customer,
) (*Order, error)
```
So for tests a simple version of the plugin can be used and for production the one that gets data from the DB.
Stateful components can be used as plugins by using the right method as argument.

## A Last Word Of Warning
The results that can be achieved with these rules can be quite astonishing and empowering.
But like any set of formal rules these can be misused and lead to ugly code bases.

The only way to make sure to stay on the good side is to keep as close as possible to the problem domain (aka 'business').
Our technical tools have to be adapted to the problems we solve and never the other way around.
A mismatch between problem and solution domain is very hard and expensive to fix later in the project.

**So please always start with the diagram first!**

These rules only tell how to translate a clean diagram into modular, maintainable and idiomatic code.
They can't help to turn a big ball of mud into a nice looking diagram.
