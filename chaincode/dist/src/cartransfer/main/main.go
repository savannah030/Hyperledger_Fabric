package main

import (
	"cartransfer/chaincode"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func main() {
	// CarTransferCC 타입의 인스턴스를 만들어 shim.Start함수에 전달
	// shim.Start 함수는 피어와의 통신 체크, 체인코드 로직 실행
	err := shim.Start(new(chaincode.CarTransferCC))
	if err != nil {
		fmt.Printf("Error in chaincode process: %s", err)
	}
}
