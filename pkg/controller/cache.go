package controller

import "sync"

type serviceExportCache struct {
	mu           sync.RWMutex              // protects serviceMap
	namespaceMap map[string]map[string]int // fromNamespace -> map[toNamespaces]1
}

func (s *serviceExportCache) set(fromNamespace, toNamespace string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	temp := make(map[string]int)
	if _, ok := s.namespaceMap[fromNamespace]; ok {
		temp = s.namespaceMap[fromNamespace]
		count := temp[toNamespace]
		count++
		temp[toNamespace] = count
	} else {
		temp[toNamespace] = 1
	}

	s.namespaceMap[fromNamespace] = temp
}

func (s *serviceExportCache) get(namespace string) (map[string]int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	toNamespace, ok := s.namespaceMap[namespace]
	return toNamespace, ok
}

func (s *serviceExportCache) toNamespaceExist(fromNamespace, toNamespace string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	toNamespaces, ok := s.namespaceMap[fromNamespace]
	if !ok {
		return false
	}

	_, ok = toNamespaces[toNamespace]
	if !ok {
		return false
	}

	return true
}

func (s *serviceExportCache) delete(fromNamespace, toNamespace string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.namespaceMap[fromNamespace]; !ok {
		return
	}

	temp := s.namespaceMap[fromNamespace]
	count := temp[toNamespace]
	count--
	temp[toNamespace] = count
	if count <= 0 {
		delete(temp, toNamespace)
	}

	if len(temp) == 0 {
		delete(s.namespaceMap, fromNamespace)
		return
	}

	s.namespaceMap[fromNamespace] = temp
}
