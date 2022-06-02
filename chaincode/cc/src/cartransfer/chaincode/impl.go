package chaincode

import (
	"github.com/hyperledger/fabric/core/chaincode/shim" //(개발일지 프로젝트 구조 참고) shim 패키지에는 체인코드 인터페이스 저장되어 있음
	pb "github.com/hyperledger/fabric/protos/peer"
)

type CarTransferCC struct {
}

// CarTransferCC 타입의 인스턴스를 받는 메서드를 초기화 하는 선언
// 예를 들어, CarTransferCC 타입의 인스턴스를 foo라고 하면
// foo.init(...) 이라는 호출을 통해 this가 foo를 가리키는 상태로 초기화 수행
func (this *CarTransferCC) Init(Stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte{})
}

func (this *CarTransferCC) Invoke(Stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte{})
}
