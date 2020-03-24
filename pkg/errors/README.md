# Error handling policy

## Handling errors from vendor functions

If you are handling an error in a vendor library, it is necessary to use `errors.Wrap()` so that you capture
a stacktrace at that time.

```go
result, err := vendorlib.DoTheirThing()
if err != nil {
    return errors.Wrap(err, "unable to do their thing")
}
```

## Handling errors from local functions

If you are handling an error from a local function, use `errors.WithMessage()` to add a message. Do not
use `errors.Wrap()` or you will overwrite the stacktrace that's already been captured.

```go
result, err := DoOurThing()
if err != nil {
    return errors.WithMessage(err, "unable to do our thing")
}
```

## Do not pass errors without adding a message

Avoid this pattern, as it makes the application harder to debug:

```go
result, err := DoOurThing()
if err != nil {
    return err
}
```

## Handling errors that happen in a loop

Unless an error in a loop should crash the application, you should log it and continue. Unhandled panic can be caught
with a deferred function:

```go
for {
    func() {
        // In case of panic, don't exit the loop, just log and continue
        defer func() {
            err := recover()
            if err != nil {
                s.log.Error(err)
            }
        }()

        result, err := DoOurThing()

        // Log returned error and stacktrace, and also add an error message
        if err != nil {
            errors.LogError(s.log, err).Error("unable to do our thing")
            continue
        }

        // ...
    }()
}
```

## Log an error that's bubbled up to the top

If the error has bubbled up to the top and should be fatal, such as errors during application initialization,
log the error using `errors.LogError()`, then stop the application with a non-zero exit code:

```go
err := rootCmd.Execute()
if err != nil {
    errors.Fatal(log, errors.Wrap(err, "app failed to run"))
}
```

or create your own error:

```go
if configValue != "expected" {
    errors.Fatal(log, errors.Errorv("unexpected config value", configValue))
}
```

## Add contextual information to an error

Use `errors.Errorv()`, `errors.Wrapv()`, and `errors.WithMessagev()` to add add contextual information to the end of
the error string

```go
foo = "bar"
foo2 = "bar2"

errors.Errorv("error string", foo)              // error string (bar)
errors.Errorv("error string", foo, foo2)        // error string (bar, bar2)
errors.Errorv("error string", logrus.Fields{    // error string (foo=bar, foo2=bar2)
    "foo": foo,
    "foo2": foo2,
})
```


