class Solution:
    def wordPattern(self, pattern: str, s: str) -> bool:
        dic = {}
        words = s.split()
        if len(pattern) != len(words) :
            return False

        for i in range(len(words)) :
            if words[i] not in dic and pattern[i] not in dic.values():
                dic[words[i]] = pattern[i]
            elif words[i] not in dic or dic[words[i]] != pattern[i]:
                return False
        return True