class Solution:
    def lengthOfLongestSubstring(self, s: str) -> int:
        max_v = 0
        now_s = ""
        for c in s :
            index = now_s.find(c)
            if index == -1:
                now_s += c
            else :
                if max_v < len(now_s) :
                    max_v = len(now_s)
                now_s = now_s[index+1:len(now_s)] + c
        return max(max_v, len(now_s))