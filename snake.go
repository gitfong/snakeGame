package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

//#include <windows.h>
//#include <conio.h>
/*
// 使用了WinAPI来移动控制台的光标
void gotoxy(int x,int y)
{
    COORD c;	//COORD是Windows API中定义的一种结构，表示一个字符在控制台屏幕上的坐标
    c.X=x,c.Y=y;
    SetConsoleCursorPosition(GetStdHandle(STD_OUTPUT_HANDLE),c);
}

// 从键盘获取一次按键，但不显示到控制台
int onKeyboard()
{
    return _getch();
}
*/
import "C" // go中可以嵌入C语言的函数

// 光标位置结构体
type loct struct {
	i, j int
}

// 随机生成一个位置，来放置食物
func randLoct() loct {
	x := rand.Int() % 10000
	y := rand.Int() % 10000
	return loct{x % 20, y % 20}
}

var (
	area      = [20][20]byte{} // 记录了蛇、食物的信息
	food      bool             // 当前是否有食物
	direction byte             // 当前蛇头移动方向
	head      loct             // 当前蛇头位置
	tail      loct             // 当前蛇尾位置
	size      int              // 当前蛇身长度
	headChar  = byte('#')
	scoreLoct = loct{22, 0} //显示分数的位置
	tipLoct   = loct{22, 2} //显示按键操作说明的位置
)

//在画布上显示c表示的字符
func draw(p loct, c byte) {
	C.gotoxy(C.int(toX(p.i)), C.int(toY(p.j)))
	fmt.Fprintf(os.Stdout, "%c", c)
}

//分数显示
func setScore(s int) {
	C.gotoxy(C.int(toX(scoreLoct.i)), C.int(toY(scoreLoct.j)))
	fmt.Fprintf(os.Stdout, "score:%d", s)
}

func setTip() {
	C.gotoxy(C.int(toX(tipLoct.i)), C.int(toY(tipLoct.j)))
	fmt.Fprint(os.Stdout, "操作提示：按方向键改变移动方向")
}

//为了显示的体验，横向显示需要间隔一个单位的横坐标，所以i要乘2
func toX(i int) int {
	return i*2 + 4
}
func toY(j int) int {
	return j + 2
}

func init() {
	// 初始化蛇头的位置和方向尾；暂时写死蛇头位置为{4,4}，移动方向向右（R）
	head, tail = loct{4, 4}, loct{4, 4}
	area[4][4] = 'H'
	direction, size = 'R', 1

	//初始显示蛇头
	draw(head, headChar)

	setScore(0)
	setTip()

	//设置随机种子
	rand.Seed(int64(time.Now().Unix()))

	//画画布前，必须把当前光标移到{0,0}位置
	C.gotoxy(0, 0)
	// 初始画布
	//由于横坐标排列比较密集，所以每两个横坐标才显示一个符号。
	//故，画布的横坐标取长度40，纵坐标取长度20)
	fmt.Fprintln(os.Stdout, `
  +-----------------------------------------+
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  +-----------------------------------------+
`)
}

func main() {
	//另起一个goroutine响应键盘事件
	go onKeyboardEvent()

	for {
		//停顿时间，其实就是控制移动速度，400毫秒移动一次
		time.Sleep(time.Millisecond * 400)

		//暂停判断
		if direction == 'P' {
			continue
		}

		// 放置食物
		if !food {
			for {
				//此循环会有性能问题，假如一直随机不到空闲位置来放置食物，可能会造成死循环
				randLoct := randLoct()
				// 值为0，表示位置目前没有显示任何符号，即为空闲位置
				if area[randLoct.i][randLoct.j] == 0 {
					area[randLoct.i][randLoct.j] = 'F' //用值F表示食物
					draw(randLoct, '$')                // 食物在画布上面的显示，用字符$表示
					food = true
					break
				}
			}

		}

		//蛇头位置保存的值direction，是用来保存当前蛇头的移动方向
		area[head.i][head.j] = direction
		oldHead := head

		switch direction {
		case 'U':
			head.j--
		case 'L':
			head.i--
		case 'R':
			head.i++
		case 'D':
			head.j++
		}

		//判断蛇头是否出界
		if head.i < 0 || head.i >= 20 || head.j < 0 || head.j >= 20 {
			dead()
			break
		}

		//获取蛇头位置的值，判断是否吃到了食物，或者撞到蛇身了
		headVal := area[head.i][head.j]
		if headVal == 'F' { // 吃到食物
			food = false

			draw(oldHead, '*')   //把蛇头变成蛇身
			draw(head, headChar) // 绘制新蛇头

			// 增加蛇的长度
			size++
			setScore(size - 1)
		} else if headVal == 0 { // 普通移动
			draw(oldHead, '*')   //把蛇头变成蛇身
			draw(head, headChar) // 绘制新蛇头

			//每个蛇尾曾经都是蛇头啊，所以蛇尾的值保存了蛇尾的移动方向
			dir := area[tail.i][tail.j]
			// 擦除蛇尾
			area[tail.i][tail.j] = 0
			draw(tail, ' ')

			// 移动蛇尾
			switch dir {
			case 'U':
				tail.j--
			case 'L':
				tail.i--
			case 'R':
				tail.i++
			case 'D':
				tail.j++
			}
		} else { // 撞到蛇身了
			dead()

			break
		}

	}

	time.Sleep(60 * time.Second)
}

//监听键盘事件(改变移动方向、暂停等)
func onKeyboardEvent() {
	for {
		switch byte(C.onKeyboard()) {
		case 72:
			if direction == 'D' {
				//跟当前方向相反，则不予处理
				continue
			}
			direction = 'U'
		case 75:
			if direction == 'R' {
				continue
			}
			direction = 'L'
		case 77:
			if direction == 'L' {
				continue
			}
			direction = 'R'
		case 80:
			if direction == 'U' {
				continue
			}
			direction = 'D'
		case 32:
			direction = 'P'
		}
	}
}

func dead() {
	C.gotoxy(0, 23)
	fmt.Fprintln(os.Stdout, "Game over!")
	//fmt.Fprintln(os.Stdout, "Press Enter key to play again; press Backspace key to exit:")
	//todo 接收通道值，判断是否再玩一次，是则continue
	C.gotoxy(0, 0)
}
