class Solution:
    def minEatingSpeed(self, piles: List[int], h: int) -> int:
        k_min = 1
        k_max = max(piles)
        
        while k_min < k_max :
            ans = (k_max+k_min) // 2
            ans_h = 0
            for p in piles :
                ans_h += (p+ans-1) // ans
            
            if ans_h <= h : 
                k_max = ans
            else :
                k_min = ans + 1

        return k_min