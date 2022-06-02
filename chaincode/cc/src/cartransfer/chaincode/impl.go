package chaincode

import (
	"cartransfer"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim" //(개발일지 프로젝트 구조 참고) shim 패키지에는 체인코드 인터페이스 저장되어 있음
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/jinzhu/inflection"
	"strings"
)

//
// Utilities
//

// Checks length of the argument
func checkLen(logger *shim.ChaincodeLogger, expected int, args []string) error {
	if len(args) < expected {
		mes := fmt.Sprintf(
			"not enough number of arguments: %d given, %d expected",
			len(args),
			expected,
		)
		logger.Warning(mes)
		return errors.New(mes)
	}
	return nil
}

//
// Chaincode interface implementation
//
type CarTransferCC struct {
}

// CarTransferCC 타입의 인스턴스를 받는 메서드를 초기화 하는 선언
// 예를 들어, CarTransferCC 타입의 인스턴스를 foo라고 하면
// foo.init(...) 이라는 호출을 통해 this가 foo를 가리키는 상태로 초기화 수행
func (this *CarTransferCC) Init(Stub shim.ChaincodeStubInterface) pb.Response {
	logger := shim.NewLogger("cartransfer")
	logger.Info("chaincode initialized")
	return shim.Success([]byte{})
}

func (this *CarTransferCC) Invoke(Stub shim.ChaincodeStubInterface) pb.Response {
	logger := shim.NewLogger("cartransfer")

	//sample of API use: show tX timestamp
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get TX timestamp: %s", err))
	}
	logger.Infof(
		"Invoke called: Tx ID = %s, timestamp = %s",
		stub.GetTxID(),
		timestamp,
	)

	var (
		fcn  string
		args []string
	)
	fcn, args = stub.GetFunctionAndParameters()
	logger.Infof("function name = %s", fcn)
	logger.Infof("args  = %s", args)

	/* 체인코드 API 함수 호출
	 * 
	 * // 소유자 관련 API
	 * 		1) 소유자 추가
	 * 		2) 소유자 목록 확인
	 * 
	 * // 자동차 관련 API
	 * 		3) 자동차 추가
	 *		4) 자동차 목록 확인
	 *      5) 소유주의 자동차 목록 확인
	 *      6) 자동차 상세 정보 확인
	 * 		7) 자동차 정보 업데이트
	 *      8) 자동차 소유권 이전

	 * // 그 외 -> 오류
	 */
	// 마샬링: 구조체->JSON | 언마샬링: JSON->구조체
	switch fcn {

		/* //////////////// 1) 소유자 추가 ////////////////
		 * 1. 입력값 개수 확인
		 * 2. 언마샬링(JSON->구조체)
		 * 3. AddOwner 함수 호출
		 * 4. shim 패키지 CarTransferCC 구조체 리턴
		 */
		case "AddOwner":
			
			// 1.
			if err := checkLen(logger, 1, args); err != nil {
				return shim.Error(err.Error())
			}

			// 2.
			owner := new(cartransfer.Owner)
			err := json.Unmarshal([]byte(args[0]), owner)
			if err != nil {
				mes := fmt.Sprintf("failed to unmarshal Owner JSON: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 3.
			err = this.AddOwner(stub, owner)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 4.
			return shim.Success([]byte{})

		
		/* //////////////// 2) 소유자 목록 확인 ////////////////
		 * 1. ListOwners 함수 호출
		 * 2. 마샬링(구조체->JSON)
		 * 3. shim 패키지 CarTransferCC 구조체 리턴
		 */
		case "ListOwners":

			// 1.
			owners, err := this.ListOwners(stub)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 2.
			b, err := json.Marshal(owners)
			if err != nil {
				mes := fmt.Sprintf("failed to marshal Owners: %s", err.Error())

				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 3.
			return shim.Success(b)


		/* //////////////// 3) 자동차 추가 ////////////////
		 * 1. 입력값 개수 확인
		 * 2. 언마샬링(JSON->구조체)
		 * 3. AddCar 함수 호출
		 * 4. shim 패키지 CarTransferCC 구조체 리턴
		 */
		case "AddCar":
			// 1.
			if err := checkLen(logger, 1, args); err != nil {
				return shim.Error(err.Error())
			}

			// 2.
			car := new(cartransfer.Car)
			err := json.Unmarshal([]byte(args[0]), car)
			if err != nil {
				mes := fmt.Sprintf("failed to unmarshal Car JSON: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 3.
			err = this.AddCar(stub, car)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 4.
			return shim.Success([]byte{})

		
		/* //////////////// 4) 자동차 목록 확인 ////////////////
		 * 1. ListCars 함수 호출
		 * 2. 마샬링(구조체->JSON)
		 * 3. shim 패키지 CarTransferCC 구조체 리턴
		 */
		case "ListCars":

			// 1.
			cars, err := this.ListCars(stub)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 2.
			b, err := json.Marshal(cars)
			if err != nil {
				mes := fmt.Sprintf("failed to marshal Cars: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 3.
			return shim.Success(b)

		
		/* //////////////// 5) 소유주의 자동차 목록 확인 ////////////////
		 * 1. 소유주 정보 언마샬링(JSON->owner)
		 * 2. ListOwnerIdCars 함수 호출
		 * 3. 여러 차 정보 마샬링(cars->JSON)
		 * 4. shim 패키지 CarTransferCC 구조체 리턴
		 */
		case "ListOwnerIdCars":

			// 1.
			var owner string
			err := json.Unmarshal([]byte(args[0]), &owner)
			if err != nil {
				mes := fmt.Sprintf("failed to unmarshal the 1st argument: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 2.
			cars, err := this.ListOwnerIdCars(stub, owner)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 3.
			b, err := json.Marshal(cars)
			if err != nil {
				mes := fmt.Sprintf("failed to marshal Cars: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 4.
			return shim.Success(b)

		
		/* //////////////// 6) 자동차 상세 정보 확인 ////////////////
		 * 1. 입력값 개수 확인
		 * 2. id 정보 언마샬링(JSON->id)
		 * 3. GetCar 함수 호출
		 * 4. 차 정보 마샬링(car->JSON)
		 * 5. shim 패키지 CarTransferCC 구조체 리턴
		 */
		case "GetCar":
			// 1.
			if err := checkLen(logger, 1, args); err != nil {
				return shim.Error(err.Error())
			}

			// 2.
			var id string
			err := json.Unmarshal([]byte(args[0]), &id)
			if err != nil {
				mes := fmt.Sprintf("failed to unmarshal the 1st argument: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 3. 
			car, err := this.GetCar(stub, id)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 4.
			b, err := json.Marshal(car)
			if err != nil {
				mes := fmt.Sprintf("failed to marshal Car: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 5.
			return shim.Success(b)


		/* //////////////// 7) 자동차 정보 업데이트 ////////////////
		 * 1. 입력값 개수 확인
		 * 2. 차 정보 언마샬링(JSON->car)
		 * 3. UpdateCar 함수 호출
		 * 4. shim 패키지 CarTransferCC 구조체 리턴
		 */	
		case "UpdateCar":
			// 1.
			if err := checkLen(logger, 1, args); err != nil {
				return shim.Error(err.Error())
			}

			// 2.
			car := new(cartransfer.Car)
			err := json.Unmarshal([]byte(args[0]), car)
			if err != nil {
				mes := fmt.Sprintf("failed to unmarshal Car JSON: %s", err.Error())
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 3.
			err = this.UpdateCar(stub, car)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 4.
			return shim.Success([]byte{})

		
		/* //////////////// 8) 자동차 소유권 이전 ////////////////
		 * 1. 입력값 개수 확인(2개)
		 * 2. carId, newOwnerId 언마샬링
		 * 3. TransferCar 함수 호출
		 * 4. shim 패키지 CarTransferCC 구조체 리턴
		 */
		case "TransferCar":
			// 1.
			if err := checkLen(logger, 2, args); err != nil {
				return shim.Error(err.Error())
			}

			// 2.
			var carId, newOwnerId string
			err := json.Unmarshal([]byte(args[0]), &carId)
			if err != nil {
				mes := fmt.Sprintf(
					"failed to unmarshal the 1st argument: %s",
					err.Error(),
				)
				logger.Warning(mes)
				return shim.Error(mes)
			}

			err = json.Unmarshal([]byte(args[1]), &newOwnerId)
			if err != nil {
				mes := fmt.Sprintf(
					"failed to unmarshal the 2nd argument: %s",
					err.Error(),
				)
				logger.Warning(mes)
				return shim.Error(mes)
			}

			// 3.
			err = this.TransferCar(stub, carId, newOwnerId)
			if err != nil {
				return shim.Error(err.Error())
			}

			// 4.
			return shim.Success([]byte{})
		}

	// 그 외 -> 오류
	mes := fmt.Sprintf("Unknown method: %s", fcn)
	logger.Warning(mes)
	return shim.Error(mes)
}

/* 함수 구현
// 소유자 관련 API
* 		1) 소유자 추가 함수
* 		2) 소유자 목록 확인 함수
* 
// 자동차 관련 API
* 		3) 자동차 추가 함수
*		4) 자동차 목록 확인 함수
*       5) 소유주의 자동차 목록 확인 함수
*       6) 자동차 상세 정보 확인 함수
* 		7) 자동차 정보 업데이트 함수
*       8) 자동차 소유권 이전 함수
*/

/* //////////////// 1) 소유자 추가 함수 ////////////////
 * 1. 소유자 이미 존재하는지 확인
 * 2. DB에 저장하기 위해 마샬링(owner->JSON)
 * 3. 복합 기본 키 생성
 * 4. DB에 저장
 */
func (this *CarTransferCC) AddOwner(stub shim.ChaincodeStubInterface, owner *cartransfer.Owner) error {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("AddOwner: Id = %s", owner.Id)

	// 1.
	found, err := this.CheckOwner(stub, owner.Id)
	if err != nil {
		return err
	}
	if found {
		mes := fmt.Sprintf("an Owner with Id = %s alerady exists", owner.Id)
		logger.Warning(mes)
		return errors.New(mes)
	}

	// 2.
	b, err := json.Marshal(owner)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// 3.
	key, err := stub.CreateCompositeKey("Owner", []string{owner.Id})
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// 4.
	err = stub.PutState(key, b)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	return nil
}

/* // 1)-1. 소유자 이미 존재하는지 확인
 * 1. 복합 기본 키 생성
 * 2. DB에서 데이터 갖고 오기
 */
func (this *CarTransferCC) CheckOwner(stub shim.ChaincodeStubInterface, id string) (bool, error) {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("CheckOwner: Id = %s", id)

	// 1.
	key, err := stub.CreateCompositeKey("Owner", []string{id})
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	// 2.
	jsonBytes, err := stub.GetState(key)
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	return jsonBytes != nil, nil
}

/* //////////////// 2) 소유자 목록 확인 함수 ////////////////
 * 1. 복합 키로 DB에서 검색
 * 2. 반복문 돌며 owners 배열에 추가
 * 3. 값이 1개일때와 2개 이상일 때로 구분하여 리턴
 */
func (this *CarTransferCC) ListOwners(stub shim.ChaincodeStubInterface) ([]*cartransfer.Owner,error) {
	logger := shim.NewLogger("cartransfer")
	logger.Info("ListOwners")

	// 1. executes a range query, which returns an iterator
	iter, err := stub.GetStateByPartialCompositeKey("Owner", []string{})
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	// will close the iterator when returned from this method
	defer iter.Close()
	owners := []*cartransfer.Owner{}

	// 2.
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		owner := new(cartransfer.Owner)
		err = json.Unmarshal(kv.Value, owner)
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		owners = append(owners, owner)
	}

	// 3.
	if len(owners) > 1 {
		logger.Infof("%d %s found", len(owners), inflection.Plural("Owner"))
	} else {
		logger.Infof("%d %s found", len(owners), "Owner")
	}
	return owners, nil
}

/* //////////////// 3) 자동차 추가 함수 ////////////////
 * 1. 복합 기본 키 생성
 * 2. 차가 이미 존재하는지 확인
 * 3. 차 검증
 * 4. DB에 저장하기 위해 마샬링(car->JSON)
 * 5. DB에 저장
 */
func (this *CarTransferCC) AddCar(stub shim.ChaincodeStubInterface, car *cartransfer.Car) error {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("AddCar: Id = %s", car.Id)

	// 1.
	key, err := stub.CreateCompositeKey("Car", []string{car.Id})
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// 2.
	found, err := this.CheckCar(stub, car.Id)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if found {
		mes := fmt.Sprintf("Car with Id = %s already exists", car.Id)
		logger.Warning(mes)
		return errors.New(mes)
	}

	// 3.
	ok, err := this.ValidateCar(stub, car)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if !ok {
		mes := "Validation of the Car failed"
		logger.Warning(mes)
		return errors.New(mes)
	}

	// 4.
	b, err := json.Marshal(car)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// 5.
	err = stub.PutState(key, b)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// returns successfully
	return nil
}

/* // 3)-2. 자동차 이미 존재하는지 확인
 * 1. 복합 기본 키 생성
 * 2. DB에서 데이터 갖고 오기
 */
func (this *CarTransferCC) CheckCar(stub shim.ChaincodeStubInterface, id string) (bool,error) {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("CheckCar: Id = %s", id)

	// 1.
	key, err := stub.CreateCompositeKey("Car", []string{id})
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	// 2.
	jsonBytes, err := stub.GetState(key)
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	// returns successfully
	return jsonBytes != nil, nil
}

/* // 3)-3. 차 검증
 * 1. 자동차의 소유주 id가 유효한지(해당 소유주가 있는지) 검증
 */
func (this *CarTransferCC) ValidateCar(stub shim.ChaincodeStubInterface,
	car *cartransfer.Car) (bool, error) {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("ValidateCar: Id = %s", car.Id)

	// 1.
	found, err := this.CheckOwner(stub, car.OwnerId)
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	// returns successfully
	return found, nil
}

/* //////////////// 4) 자동차 목록 확인 함수 ////////////////
 * 1. 복합 키로 DB에서 검색
 * 2. 반복문 돌며 cars 배열에 추가
 * 3. 값이 1개일때와 2개 이상일 때로 구분하여 리턴
 */
func (this *CarTransferCC) ListCars(stub shim.ChaincodeStubInterface) ([]*cartransfer.Car, error) {
	logger := shim.NewLogger("cartransfer")
	logger.Info("ListCars")

	// 1. executes a range query, which returns an iterator
	iter, err := stub.GetStateByPartialCompositeKey("Car", []string{})
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	// will close the iterator when returned from this method
	defer iter.Close()

	// 2.
	cars := []*cartransfer.Car{}
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		car := new(cartransfer.Car)
		err = json.Unmarshal(kv.Value, car)
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		cars = append(cars, car)
	}

	// 3. returns successfully
	if len(cars) > 1 {
		logger.Infof("%d %s found", len(cars), inflection.Plural("Car"))
	} else {
		logger.Infof("%d %s found", len(cars), "Car")
	}
	return cars, nil
}

/* //////////////// 5) 소유주의 자동차 목록 확인 함수 ////////////////
 * 1. 복합 키로 DB에서 검색
 * 2. 반복문 돌며 cars 배열에 추가
 * 3. 값이 1개일때와 2개 이상일 때로 구분하여 리턴
 */
func (this *CarTransferCC) ListOwnerIdCars(stub shim.ChaincodeStubInterface, ownerId string) ([]*cartransfer.Car,error) {
	logger := shim.NewLogger("cartransfer")
	logger.Info("ListOwnerCars")

	// 1. executes a range query, which returns an iterator
	iter, err := stub.GetStateByPartialCompositeKey("Car", []string{})
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	// will close the iterator when returned from this method
	defer iter.Close()

	// 2. 
	cars := []*cartransfer.Car{}
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		car := new(cartransfer.Car)
		err = json.Unmarshal(kv.Value, car)
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		if strings.Index(ownerId, "admin") != -1 {
			cars = append(cars, car)
		} else {
			if car.OwnerId == ownerId {
				cars = append(cars, car)
			}

		}

	}

	// 3. returns successfully
	if len(cars) > 1 {
		logger.Infof("%d %s found", len(cars), inflection.Plural("Car"))
	} else {
		logger.Infof("%d %s found", len(cars), "Car")
	}
	return cars, nil
}

/* //////////////// 6) 자동차 상세 정보 확인 함수 ////////////////
 * 1. 복합 기본 키 생성
 * 2. DB에서 데이터 갖고 오기
 * 3. 언마샬링(JSON->car)
 */
func (this *CarTransferCC) GetCar(stub shim.ChaincodeStubInterface,
	id string) (*cartransfer.Car, error) {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("GetCar: Id = %s", id)

	// 1.
	key, err := stub.CreateCompositeKey("Car", []string{id})
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	// 2.
	jsonBytes, err := stub.GetState(key)
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}
	if jsonBytes == nil {
		mes := fmt.Sprintf("Car with Id = %s was not found", id)
		logger.Warning(mes)
		return nil, errors.New(mes)
	}

	// 3.
	car := new(cartransfer.Car)
	err = json.Unmarshal(jsonBytes, car)
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	// returns successfully
	return car, nil
}

/* //////////////// 7) 자동차 정보 업데이트 함수 ////////////////
 * 1. 자동차 이미 존재하는지 확인
 * 2. 자동차 검증
 * 3. 복합 키 생성
 * 4. DB에 저장하기 위해 마샬링(car->JSON)
 * 5. DB에 저장
 */
func (this *CarTransferCC) UpdateCar(stub shim.ChaincodeStubInterface,
	car *cartransfer.Car) error {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("UpdateCar: car = %+v", car)

	// 1.
	found, err := this.CheckCar(stub, car.Id)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if !found {
		mes := fmt.Sprintf("Car with Id = %s does not exist", car.Id)
		logger.Warning(mes)
		return errors.New(mes)
	}

	// 2.
	ok, err := this.ValidateCar(stub, car)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if !ok {
		mes := "Validation of the Car failed"
		logger.Warning(mes)
		return errors.New(mes)
	}

	// 3.
	key, err := stub.CreateCompositeKey("Car", []string{car.Id})
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// 4.
	b, err := json.Marshal(car)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// 5.
	err = stub.PutState(key, b)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// returns successfully
	return nil
}

/* //////////////// 8) 자동차 소유권 이전 함수 ////////////////
 * 1. 자동차 객체 가져오기
 * 2. 소유주 id 업데이트
 * 3. 자동차 정보 업데이트( 7) 자동차 정보 업데이트 함수 호출)
 */
func (this *CarTransferCC) TransferCar(stub shim.ChaincodeStubInterface, carId string, newOwnerId string) error {
	logger := shim.NewLogger("cartransfer")
	logger.Infof("TransferCar: Car Id = %s, new Owner Id = %s", carId, newOwnerId)

	// 1.
	car, err := this.GetCar(stub, carId)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// 2.
	car.OwnerId = newOwnerId

	// 3.
	err = this.UpdateCar(stub, car)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	// returns successfully
	return nil
}