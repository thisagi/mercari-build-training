class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        A = headA
        B = headB

        while A != B:
            if A is None :
                A = headB
            else :
                A = A.next
                
            if B is None :
                B = headA
            else :
                B = B.next
        return A