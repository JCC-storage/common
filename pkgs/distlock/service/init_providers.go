package service

import (
	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/pkgs/distlock/lockprovider"
	"gitlink.org.cn/cloudream/common/pkgs/distlock/service/internal"
	"gitlink.org.cn/cloudream/common/pkgs/trie"
)

func initProviders(providers *internal.ProvidersActor) {
	initMetadataLockProviders(providers)

	initIPFSLockProviders(providers)

	initStorageLockProviders(providers)
}

func initMetadataLockProviders(providers *internal.ProvidersActor) {
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "Node")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "Storage")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "User")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "UserBucket")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "UserNode")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "UserStorage")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "Bucket")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "Object")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "ObjectRep")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "ObjectBlock")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "Cache")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "StorageObject")
	providers.AddProvider(lockprovider.NewMetadataLock(), distlock.METADATA_LOCK_PATH_PREFIX, "Location")
}

func initIPFSLockProviders(providers *internal.ProvidersActor) {
	providers.AddProvider(lockprovider.NewIPFSLock(), distlock.IPFS_LOCK_PATH_PREFIX, trie.WORD_ANY)
}

func initStorageLockProviders(providers *internal.ProvidersActor) {
	providers.AddProvider(lockprovider.NewStorageLock(), distlock.STORAGE_LOCK_PATH_PREFIX, trie.WORD_ANY)
}
