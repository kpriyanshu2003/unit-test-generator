#include <gtest/gtest.h>
#include "example.h"
#include <stdexcept>
#include <cmath>

TEST(CalculatorTests, addPositive) {
    Calculator calc;
    EXPECT_EQ(calc.add(5, 3), 8);
}

TEST(CalculatorTests, addNegative) {
    Calculator calc;
    EXPECT_EQ(calc.add(-5, -3), -8);
}

TEST(CalculatorTests, subtractPositive) {
    Calculator calc;
    EXPECT_EQ(calc.subtract(10, 4), 6);
}

TEST(CalculatorTests, subtractNegative) {
    Calculator calc;
    EXPECT_EQ(calc.subtract(-10, -4), -6);
}

int main(int argc, char **argv) {
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}