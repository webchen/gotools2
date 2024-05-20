package fun

import (
	"time"
)

// AddDate 时间添加，处理日期头尾日BUG的问题
func AddDate(t time.Time, year, month, day int) time.Time {
	//先跳到目标月的1号
	targetDate := t.AddDate(year, month, -t.Day()+1)
	//获取目标月的临界值
	targetDay := targetDate.AddDate(0, 1, -1).Day()
	//对比临界值与源日期值，取最小的值
	if targetDay > t.Day() {
		targetDay = t.Day()
	}
	//最后用目标月的1号加上目标值和入参的天数
	targetDate = targetDate.AddDate(0, 0, targetDay-1+day)
	return targetDate
}
