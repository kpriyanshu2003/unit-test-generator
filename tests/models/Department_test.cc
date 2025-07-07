#include <gtest/gtest.h>
#include <cmath>
#include <stdexcept>
#include <json/json.h>

#include "Department.h"

using namespace drogon::orm;
using drogon_model::org_chart::Department;

TEST(DepartmentTest, DefaultConstructor)
{
    Department department;
    EXPECT_EQ(department.getId(), nullptr);
    EXPECT_EQ(department.getValueOfId(), 0);
}

TEST(DepartmentTest, JsonConstructor)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["name"] = "Department";
    Department department(pJson);
    EXPECT_EQ(department.getValueOfId(), 1);
    EXPECT_EQ(department.getValueOfName(), "Department");
}

TEST(DepartmentTest, IdGetter)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["name"] = "Department";
    Department department(pJson);
    const int32_t &value = department.getValueOfId();
    EXPECT_EQ(value, 1);
}

TEST(DepartmentTest, IdSetter)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["name"] = "Department";
    Department department(pJson);
    department.setId(2);
    const int32_t &value = department.getValueOfId();
    EXPECT_EQ(value, 2);
}

TEST(DepartmentTest, NameGetter)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["name"] = "Department";
    Department department(pJson);
    const std::string &value = department.getValueOfName();
    EXPECT_EQ(value, "Department");
}

TEST(DepartmentTest, NameSetter)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["name"] = "Department";
    Department department(pJson);
    department.setName("New Department");
    const std::string &value = department.getValueOfName();
    EXPECT_EQ(value, "New Department");
}
