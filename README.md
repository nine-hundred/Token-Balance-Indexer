# [온블록 BE 과제] 블록체인 이벤트 기반 Token Balance Indexer 개발

## 개요
블록체인 네트워크에서 발생하는 토큰 이벤트를 실시간으로 인덱싱하여 계정별 토큰 잔액을 추적하는 MSA 기반 시스템입니다.

### 데이터 흐름
TX-Indexer(GraphQL) → Block-Synchronizer → SQS → Event-Processing → Postgres DB → Balance-API

## 실행 방법

### 1. 인프라 환경 셋업

````bash
# 필요한 환경 시작
make env-up

# 환경 종료
make env-down
````

### 2. 서비스 실행

````bash
# block-synchronizer 실행
make run-synchronizer

# event-processor 실행
make run-processor

# balance-api 실행
make run-api

# 이 외 사용 가능한 명령어 확인
make help
````

## 설계 의도

### 디렉토리/패키지 구조
디렉토리/패키지 구조는 [공식 문서](https://github.com/golang-standards/project-layout)를 따르는 것을 기준으로 삼았습니다.

````
cmd/ # 메인 애플리케이션들
├── block-Synchronizer/
├── event-processor/
└── balance-api/
internal/ # 각 서버 내부에서만 사용하는 코드
├── config/
├── consumer/
├── handler/
├── repository/
├── response/
├── service/
└── tx-indexer/
pkg/ # 다른 패키지에서도 사용 가능한 코드 묶음
├── caching/ 
├── messaging/ 
└── model/
````

### 인터페이스 기반 확장성 고려
처음에는 서비스에서 `*redis.Client` 를 직접 사용할지 고민했습니다.
````
// 직접 사용
type EventProcessor struct {
	caching         *redis.Client
}

// 레퍼 사용
type EventProcessor struct {
	caching         caching.Caching
}
````

단순히 redis를 감싼 것처럼 보이지만, 향후 로컬 캐시나 memcached 등 추후 확장을 위해 Caching 인터페이스로 분리하였습니다.
````
type Caching interface {
	IncrBy(ctx context.Context, key string, value int64) error
	DecrBy(ctx context.Context, key string, value int64) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
}
````
SQS도 AWS에 종속되지 않도록 고려하였습니다.
````
struct MessageObject{} ...

func (p SQSClient) PublishMessage(ctx context.Context, event interface{}) error {}

func (p SQSClient) ReceiveMessage(ctx context.Context) (message MessageObject, err error) {}
````

### Block-Synchronizer
백필의 배치 사이즈는 5,000으로 설정했습니다.

배치 사이즈 결정 과정:
1. 초기 접근은 서버 다운타임 동안 발생한 모든 블록을 한 번의 요청으로 처리
2. 최대 10,000개 블록까지만 조회 가능하다는 GraphQL 제약을 확인
3. 로컬 환경에서 다양한 배치 사이즈 실험
4. 5,000개가 지연 없이 가장 빠른 처리 속도 보여줌

개선 사항: 운영 환경에서 모니터링을 통해 추가 테스트 필요함.

### Event-Processing
가장 고민을 많이 했던 부분입니다.
초기에는 높은 TPS 가운데 데이터의 일관성과 정합성을 고려하여 아키텍쳐를 고민했습니다.
````
1. FIFO SQS + MessageGroupId로 이벤트 순서 보장
2. 계층 분리
   1. Postgres: 모든 이벤트를 영구 저장(복구용)
   2. Redis: 실시간 잔액 계산 및 캐싱(성능 최적화)
      1. DB-Redis 초(분)당 동기화 배치 작업 추가.
3. 장기 운영 고려
   1. 잔액 데이터에 대한 영속성을 고려하여 DB에 일간/월간 스냅샷 배치 작업을 /cmd/batch 에 추가
````

그러나 GraphQL을 모니터링해보니 블록 height값 증가량과 이벤트가 많지 않다는 것을 알게 되었습니다.

오버엔지니어링이란 것을 깨달은 후, 설계를 단순화하였습니다.

````
Redis 제외:
 - 실제 TPS를 고려해본 결과 Postgres 단독으로 충분히 해결 가능(트랜잭션)
 - (transaction_hash, tx_event_index)` 조합으로 멱등성 보장
 - 인프라 복잡성 제거

Standard SQS 전환
 - FIFO 제약 해제로 구현 단순화
 - 원자적 upsert로 동시성 문제 해결
````

### Balance-API
라우팅 설계 고려 사항

문제: `tokenPath` 파라미터에 슬래시(/)가 포함된 경로가 포함되어야 함.

ex) `gno.land/r/gnoswap/v1/test_token/bar` , `gno.land/r/gnoswap/v1/test_token/foo` ..

base64로 변경이나 쿼리파라미터를 활용한 api명세 변경은 불가능하다고 판단하였습니다.
따라서 와일드카드로 라우팅을 분기하였습니다.
````
tokenGroup.GET("/*wildcard", func(c *gin.Context) {
		wildcard := c.Param("wildcard")
		if wildcard == "/balances" {
			handler.GetTokenBalances(c)
		} else if wildcard == "/transfer-history" {
			handler.GetTokenTransferHistory(c)
		} else {
			handler.GetTokenPathBalances(c)
		}
	})
````

### 저장소 설계

데이터베이스 스키마는 `schema.sql`에 정의되어 있습니다.

주요 테이블:

| 테이블명           | 역할                |
|----------------|-------------------|
| `blocks`       | 블록 정보 저장         |
| `transactions` | 트랜잭션 내역 기록       |
| `token_events` | 파싱된 토큰 이벤트 기록   |
| `balances`     | 계산된 토큰 잔액       |

## 개선 사항 및 한계
아래 사항은 시간 제약과 우선 순위에 밀려 구현하지 못한 부분입니다.
- **에러 타입 체계화**: 현재 기본 에러 타입 사용, 추후 도메인/레이어 별 커스텀 에러 타입 설계 필요
- **SQS Long Polling**: 현재는 SQS의 WaitTimeout을 적용하지 않음
- **메트릭 및 모니터링**: 헬스체크, 성능 지표 수집 등 운영 관점의 기능 미구현

### 운영 환경 고려사항
- 배치 사이즈 최적화를 위한 추가 성능 테스트 필요
- 장애 복구 시나리오별 테스트 필요