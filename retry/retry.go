package retry

import (
	"time"
)

type RetryError struct{
	RequireRetry bool
	Error error
}


type RetryFunction func() (interface{},*RetryError)

type timeoutError struct {
	duration time.Duration
}

func (e *timeoutError) Error() string {
	return "Timeout after " + e.duration.String()
}



func Do(retryableFunction RetryFunction) (interface{},error)  {
	var retryTime time.Duration
	startActionTime := time.Now()
	retryTime=time.Duration(100) * time.Millisecond
	for {
		value,retryError := retryableFunction()
		if retryError != nil {
			if retryError.RequireRetry {
				retryTime=retryTime*2
				if retryTime>(time.Duration(15) * time.Second) {
					retryTime=time.Duration(15) * time.Second
				}
				elapsed := time.Now().Sub(startActionTime)

				if(elapsed.Hours()>24){
					return nil,&timeoutError{
						duration:elapsed,
					}
				}
				time.Sleep(retryTime)
			}else{
				return nil,retryError.Error
			}
		}else{
			return value,nil
		}
	}
}
