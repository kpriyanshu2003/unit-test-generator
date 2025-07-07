#include <gtest/gtest.h>
#include <cmath>
#include <stdexcept>
#include <drogon/orm/Result.h>
#include <drogon/orm/Row.h>
#include <drogon/orm/Field.h>
#include <drogon/orm/SqlBinder.h>
#include <drogon/orm/Mapper.h>
#include <trantor/utils/Date.h>
#include <trantor/utils/Logger.h>
#include <json/json.h>
#include <string>
#include <memory>
#include <vector>
#include <tuple>
#include <stdint.h>
#include <iostream>
#include "PersonInfo.h"

using namespace drogon::orm;
using namespace drogon_model::org_chart;

TEST(PersonInfoTest, ConstructorTest)
{
    // Test default constructor
    PersonInfo personInfo1;
    EXPECT_EQ(personInfo1.getId(), nullptr);
    EXPECT_EQ(personInfo1.getValueOfId(), 0);

    // Test another default constructor instance
    PersonInfo personInfo2;
    EXPECT_EQ(personInfo2.getJobTitle(), nullptr);
    EXPECT_EQ(personInfo2.getValueOfJobTitle(), "");
}

TEST(PersonInfoTest, GetterSetterTest)
{
    PersonInfo personInfo;

    // Test ID field
    EXPECT_EQ(personInfo.getId(), nullptr);
    EXPECT_EQ(personInfo.getValueOfId(), 0);

    // Test all other fields
    EXPECT_EQ(personInfo.getJobId(), nullptr);
    EXPECT_EQ(personInfo.getValueOfJobId(), 0);

    EXPECT_EQ(personInfo.getDepartmentId(), nullptr);
    EXPECT_EQ(personInfo.getValueOfDepartmentId(), 0);

    EXPECT_EQ(personInfo.getManagerId(), nullptr);
    EXPECT_EQ(personInfo.getValueOfManagerId(), 0);

    EXPECT_EQ(personInfo.getJobTitle(), nullptr);
    EXPECT_EQ(personInfo.getValueOfJobTitle(), "");

    EXPECT_EQ(personInfo.getDepartmentName(), nullptr);
    EXPECT_EQ(personInfo.getValueOfDepartmentName(), "");

    EXPECT_EQ(personInfo.getManagerFullName(), nullptr);
    EXPECT_EQ(personInfo.getValueOfManagerFullName(), "");

    EXPECT_EQ(personInfo.getFirstName(), nullptr);
    EXPECT_EQ(personInfo.getValueOfFirstName(), "");

    EXPECT_EQ(personInfo.getLastName(), nullptr);
    EXPECT_EQ(personInfo.getValueOfLastName(), "");

    EXPECT_EQ(personInfo.getHireDate(), nullptr);
}