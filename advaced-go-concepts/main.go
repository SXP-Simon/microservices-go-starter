package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type contextKey string

var UserIdKey contextKey = "UserID"

var (
	ErrorNotImplemented = errors.New("not implemented")
	ErrorTruckNotFound  = errors.New("not found")
)

type Truck interface {
	LoadCargo() error
	UnloadCargo() error
}
type NormalTruck struct {
	id    string
	cargo int
}
type ElectricTruck struct {
	id      string
	cargo   int
	battery int
}

func (t *NormalTruck) LoadCargo() error {
	t.cargo += 2
	return errors.New("Some error")
}
func (t *NormalTruck) UnloadCargo() error {
	t.cargo = 0
	return nil
}

func (e *ElectricTruck) LoadCargo() error {
	e.cargo += 2
	e.battery += 2
	return nil
}

func (e *ElectricTruck) UnloadCargo() error {
	e.cargo = 0
	e.battery = 0
	return nil
}

func processTruck(ctx context.Context, truck Truck) error {
	fmt.Printf("Started processing truck %+v\n", truck)

	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	time.Sleep(time.Second)

	if err := truck.LoadCargo(); err != nil {
		return fmt.Errorf("Error loading cargo: %w \n", err)
	}

	fmt.Printf("Finished processing truck %+v\n", truck)
	return nil
}

func processFleet(ctx context.Context, fleet []Truck) error {
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(fleet))

	for _, t := range fleet {
		wg.Add(1)
		go func(t Truck) {
			if err := processTruck(ctx, t); err != nil {
				log.Println(err)
				errorsChan <- err
			}
			wg.Done()
		}(t)
	}

	wg.Wait()
	close(errorsChan)

	var errs []error
	for err := range errorsChan {
		log.Printf("Error processing truck: %v\n", err)
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("fleet processing had %v errors\n", len(errs))
	}
	return nil
}

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIdKey, 233)

	fleet := []Truck{
		&NormalTruck{id: "NT1", cargo: 0},
		&ElectricTruck{id: "ET1", cargo: 0, battery: 100},
		&NormalTruck{id: "NT2", cargo: 0},
		&ElectricTruck{id: "ET2", cargo: 0, battery: 100},
	}
	if err := processFleet(ctx, fleet); err != nil {
		log.Fatal(err)
	}
}
