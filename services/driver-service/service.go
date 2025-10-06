package main

import (
	"math/rand"
	"strings"
	math "math/rand/v2"
	pb "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/util"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type driverInMap struct {
	Driver *pb.Driver
	// Index int
	// TODO: route
}

type Service struct {
	drivers []*driverInMap
	mu      sync.RWMutex
}

func NewService() *Service {
	return &Service{
		drivers: make([]*driverInMap, 0),
	}
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIndex := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIndex]

	randomPlate := GenerateRandomPlate()
	randomAvatar := util.GetRandomAvatar(randomIndex)

	// we can ignore this property for now, but it must be sent to the frontend.
	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver := &pb.Driver{
		Geohash:        geohash,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:           "Lando Norris",
		Id:             driverId,
		PackageSlug:    packageSlug,
		ProfilePicture: randomAvatar,
		CarPlate:       randomPlate,
	}

	s.drivers = append(s.drivers, &driverInMap{
		Driver: driver,
	})

	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, driver := range s.drivers {
		if driverId == driver.Driver.Id {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)

		}
	}

}

// FindNearbyDrivers 查找附近的司机
func (s *Service) FindNearbyDrivers(pickupGeohash, packageSlug string) []*driverInMap {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nearbyDrivers []*driverInMap
	
	// 获取geohash的前缀（精度为7，大约1.5km范围）
	prefix := pickupGeohash[:7]
	
	for _, driver := range s.drivers {
		// 检查司机类型是否匹配
		if driver.Driver.PackageSlug != packageSlug {
			continue
		}
		
		// 检查司机是否在附近（使用geohash前缀匹配）
		if strings.HasPrefix(driver.Driver.Geohash, prefix) {
			nearbyDrivers = append(nearbyDrivers, driver)
		}
	}
	
	// 随机打乱司机顺序，避免总是选择相同的司机
	rand.Shuffle(len(nearbyDrivers), func(i, j int) {
		nearbyDrivers[i], nearbyDrivers[j] = nearbyDrivers[j], nearbyDrivers[i]
	})
	
	// 最多返回3个司机
	if len(nearbyDrivers) > 3 {
		nearbyDrivers = nearbyDrivers[:3]
	}
	
	return nearbyDrivers
}

// UpdateDriverLocation 更新司机位置
func (s *Service) UpdateDriverLocation(driverID string, location *pb.Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for _, driver := range s.drivers {
		if driver.Driver.Id == driverID {
			// 更新位置
			driver.Driver.Location = location
			// 更新geohash
			driver.Driver.Geohash = geohash.Encode(location.Latitude, location.Longitude)
			break
		}
	}
}
