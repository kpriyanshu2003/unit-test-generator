#include <gtest/gtest.h>
#include <cmath>
#include <stdexcept>

TEST(CalculatorTest, AddPositive) {
    Calculator calc;
    EXPECT_EQ(calc.add(5, 3), 8);
}

TEST(CalculatorTest, AddNegative) {
    Calculator calc;
    EXPECT_EQ(calc.add(-2, -1), -3);
}

TEST(CalculatorTest, SubtractPositive) {
    Calculator calc;
    EXPECT_EQ(calc.subtract(10, 4), 6);
}

TEST(CalculatorTest, SubtractNegative) {
    Calculator calc;
    EXPECT_EQ(calc.subtract(-5, -8), 3);
}