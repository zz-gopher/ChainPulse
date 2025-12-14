package main

import (
	"ChainPulse/decoder"
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 1. 连接节点
	client, err := ethclient.Dial("https://eth.llamarpc.com")
	if err != nil {
		log.Fatal(err)
	}

	transferSign := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	// 2. 模拟数据库里的游标 (生产环境从 DB 读取)
	// 假设我们上次处理到了这个区块
	var currentBlock int64 = 24009270

	// 设置我们要监听的合约地址 (比如 USDT)
	contractAddr := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")

	fmt.Println("🚀 HexHunter 扫描器启动...")

	// 3. 死循环：永不停止的扫描
	for {
		// A. 获取链上最新高度
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			fmt.Println("节点连接失败，重试中...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		chainHead := header.Number.Int64()

		// B. 判断有没有新区块
		if currentBlock >= chainHead {
			// 还没有新区块，休息一下
			fmt.Printf("⏳ 等待新区块... (当前: %d)\n", currentBlock)
			time.Sleep(12 * time.Second) // 以太坊每12秒一个块，BSC 3秒
			continue
		}

		// C. 计算这一轮要扫的范围 (Batch Processing)
		// 为了防止一次查太多导致节点报错，我们一次只扫 10 个块
		toBlock := currentBlock + 10
		if toBlock > chainHead {
			toBlock = chainHead
		}

		fmt.Printf("🔍 正在扫描区块范围: [%d -> %d]\n", currentBlock+1, toBlock)

		// D. 构建查询
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(currentBlock + 1),
			ToBlock:   big.NewInt(toBlock),
			// 监听的合约地址
			Addresses: []common.Address{contractAddr},
			// 监听的事件
			Topics: [][]common.Hash{
				{transferSign},
			},
		}

		// E. 抓取日志
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Println("抓取日志失败:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// F. 处理每一条日志 (你的解码逻辑放在这！)
		for _, vLog := range logs {
			fmt.Printf("🔥 发现事件！在区块 #%d, TxHash: %s\n", vLog.BlockNumber, vLog.TxHash.Hex())

			// ===================================
			// [在此处插入你之前的 Decoder 代码]
			// 1. 解析 Topics -> From/To
			// 2. 解析 Data -> Value
			// 3. Insert into Database
			// ===================================
			transferData, err := decoder.TransferParse(vLog.Topics, common.Bytes2Hex(vLog.Data))
			if err != nil {
				// ⚠️ 建议用 Println 而不是 Fatal，防止单条脏数据导致程序崩溃
				log.Println("解码失败，跳过:", err)
				continue
			}
			fmt.Println("------------------------------------------------")
			fmt.Printf("🧱 区块高度 : %d\n", vLog.BlockNumber)
			fmt.Printf("🔗 交易 Hash: %s\n", vLog.TxHash.Hex())
			fmt.Printf("📤 发送方   : %s\n", transferData.Sender.Hex())
			fmt.Printf("📥 接收方   : %s\n", transferData.To.Hex())
			fmt.Printf("💰 原始金额 : %s\n", transferData.Value.String())
		}

		// G. 更新游标 (这一步至关重要！)
		// 只有确认上面的数据都入库了，才更新这个数字
		currentBlock = toBlock
		// TODO: db.Save("last_block", currentBlock)
	}
}
