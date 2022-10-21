package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types", func() {
	Context("Stack", func() {
		var stack Stack[int]

		BeforeEach(func() {
			stack = Stack[int]{}
		})
		When("a stack is created", func() {
			It("should be empty", func() {
				Expect(stack.Empty()).To(BeTrue())
			})
		})
		When("an element is pushed", func() {
			BeforeEach(func() {
				stack.Push(10)
			})

			It("should not be empty", func() {
				Expect(stack.Empty()).To(BeFalse())
			})
			It("should return that element when popped and be empty", func() {
				e, err := stack.Pop()
				Expect(err).ToNot(HaveOccurred())
				Expect(e).To(Equal(10))
				Expect(stack.Empty()).To(BeTrue())
			})
		})
		When("popping an empty stack", func() {
			It("should not break", func() {
				_, err := stack.Pop()
				Expect(err).To(Equal(StackEmptyError))
			})
		})
	})
})
