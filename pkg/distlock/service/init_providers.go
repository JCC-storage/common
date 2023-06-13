package service

import (
	"gitlink.org.cn/cloudream/common/pkg/distlock/lockprovider"
	"gitlink.org.cn/cloudream/common/pkg/distlock/service/internal"
	"gitlink.org.cn/cloudream/common/pkg/trie"
)

func initProviders(providers *internal.ProvidersActor) {
	initMetadataLockProviders(providers)

	initIPFSLockProviders(providers)

	initStorageLockProviders(providers)
}

func initMetadataLockProviders(providers *internal.ProvidersActor) {
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "Node")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "Storage")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "User")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "UserBucket")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "UserNode")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "UserStorage")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "Bucket")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "Object")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "ObjectRep")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "ObjectBlock")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "Cache")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "StorageObject")
	providers.AddProvider(lockprovider.NewMetadataLock(), "Metadata", "Location")
}

func initIPFSLockProviders(providers *internal.ProvidersActor) {
	providers.AddProvider(lockprovider.NewIPFSLock(), "IPFS", trie.WORD_ANY)
}

func initStorageLockProviders(providers *internal.ProvidersActor) {
	providers.AddProvider(lockprovider.NewStorageLock(), "Storage", trie.WORD_ANY)
}
