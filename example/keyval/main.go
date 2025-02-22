//
// Copyright (C) 2019 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/dynamo
//

package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fogfish/curie"
	"github.com/fogfish/dynamo"
)

//
// Person type demonstrates composition of core type with db one
type Person struct {
	Org     curie.IRI `dynamodbav:"prefix,omitempty"`
	ID      curie.IRI `dynamodbav:"suffix,omitempty"`
	Name    string    `dynamodbav:"name,omitempty"`
	Age     int       `dynamodbav:"age,omitempty"`
	Address string    `dynamodbav:"address,omitempty"`
}

var codec = dynamo.Struct(Person{}).Codec("Org", "ID")

//
func (p Person) Identity() (string, string) { return p.Org.String(), p.ID.String() }

//
func (p Person) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	type tStruct Person
	return dynamo.Encode(av, tStruct(p),
		codec.Encode(dynamo.IRI(p.Org), dynamo.IRI(p.ID)),
	)
}

//
func (p *Person) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	type tStruct *Person
	return dynamo.Decode(av, tStruct(p),
		codec.Decode((*dynamo.IRI)(&p.Org), (*dynamo.IRI)(&p.ID)),
	)
}

//
type Persons []Person

// Join ...
func (seq *Persons) Join(gen dynamo.Gen) error {
	val := Person{}
	if fail := gen.To(&val); fail != nil {
		return fail
	}
	*seq = append(*seq, val)
	return nil
}

// KeyVal is type synonym
type KeyVal dynamo.KeyValNoContext

//
//
func main() {
	db := dynamo.NewKeyValContextDefault(dynamo.Must(dynamo.New(os.Args[1])))

	examplePut(db)
	exampleGet(db)
	exampleUpdate(db)
	exampleMatch(db)
	exampleRemove(db)
}

const n = 5

func examplePut(db KeyVal) {
	for i := 0; i < n; i++ {
		val := &Person{
			Org:     curie.New("test:"),
			ID:      curie.New("person:%d", i),
			Name:    "Verner Pleishner",
			Age:     64,
			Address: "Blumenstrasse 14, Berne, 3013",
		}
		err := db.Put(val)

		fmt.Println("=[ put ]=> ", either(err, val))
	}
}

func exampleGet(db KeyVal) {
	for i := 0; i < n; i++ {
		val := &Person{
			Org: curie.New("test:"),
			ID:  curie.New("person:%d", i),
		}
		switch err := db.Get(val).(type) {
		case nil:
			fmt.Printf("=[ get ]=> %+v\n", val)
		case dynamo.NotFound:
			fmt.Printf("=[ get ]=> Not found: (%v, %v)\n", val.Org, val.ID)
		default:
			fmt.Printf("=[ get ]=> Fail: %v\n", err)
		}
	}
}

func exampleUpdate(db KeyVal) {
	for i := 0; i < n; i++ {
		val := &Person{
			Org:     curie.New("test:"),
			ID:      curie.New("person:%d", i),
			Address: "Viktoriastrasse 37, Berne, 3013",
		}
		err := db.Update(val)

		fmt.Printf("=[ update ]=> %+v\n", either(err, val))
	}
}

func exampleMatch(db KeyVal) {
	seq := Persons{}
	err := db.Match(Person{Org: curie.New("test:")}).FMap(seq.Join)

	if err == nil {
		fmt.Printf("=[ match ]=> %+v\n", seq)
	} else {
		fmt.Printf("=[ match ]=> %v\n", err)
	}
}

func exampleRemove(db KeyVal) {
	for i := 0; i < n; i++ {
		val := &Person{
			Org: curie.New("test:"),
			ID:  curie.New("person:%d", i),
		}
		err := db.Remove(val)

		fmt.Println("=[ remove ]=> ", either(err, val))
	}
}

func either(e error, x interface{}) interface{} {
	if e != nil {
		return e
	}
	return x
}
