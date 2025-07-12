#include "gtest/gtest.h"
#include "example.h"
#include <stdexcept>
#include <cmath>

TEST(CalculatorTest, AddPositive) {
    Calculator calculator;
    EXPECT_EQ(calculator.add(5, 3), 8);
}

TEST(CalculatorTest, AddNegative) {
    Calculator calculator;
    EXPECT_EQ(calculator.add(-5, -3), -8);
}

TEST(CalculatorTest, SubtractPositive) {
    Calculator calculator;
    EXPECT_EQ(calculator.subtract(10, 4), 6);
}

TEST(CalculatorTest, SubtractNegative) {
    Calculator calculator;
    EXPECT_EQ(calculator.subtract(-10, -4), -6);
}

int main(int argc, char **argv) {
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}