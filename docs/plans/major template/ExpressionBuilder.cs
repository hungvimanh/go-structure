using Microsoft.EntityFrameworkCore.Query;
using Microsoft.EntityFrameworkCore.Query.SqlExpressions;
using System.Reflection;

namespace CHECKIN.Common.FilterExtensions
{
    public static class ExpressionBuilder
    {
        public static bool Contains(this Enum source, Enum destination)
        {
            long sourceValue = Convert.ToInt64(source);
            long destValue = Convert.ToInt64(destination);
            return (sourceValue & destValue) == destValue;
        }

        public static string ToSql<TEntity>(this IQueryable<TEntity> query) where TEntity : class
        {
            var enumerator = query.Provider.Execute<IEnumerable<TEntity>>(query.Expression).GetEnumerator();
            var relationalCommandCache = enumerator.Private("_relationalCommandCache");
            var selectExpression = relationalCommandCache.Private<SelectExpression>("_selectExpression");
            var factory = relationalCommandCache.Private<IQuerySqlGeneratorFactory>("_querySqlGeneratorFactory");

            var sqlGenerator = factory.Create();
            var command = sqlGenerator.GetCommand(selectExpression);

            string sql = command.CommandText;
            return sql;
        }

        public static object Private(this object obj, string privateField) => obj?.GetType().GetField(privateField, BindingFlags.Instance | BindingFlags.NonPublic)?.GetValue(obj);

        public static T Private<T>(this object obj, string privateField) => (T)obj?.GetType().GetField(privateField, BindingFlags.Instance | BindingFlags.NonPublic)?.GetValue(obj);

        public static Expression<Func<TSource, bool>> BuildPredicate<TSource, TKey, TValue>(Expression<Func<TSource, TKey>> propertyName, string comparison, List<TValue> values)
        {
            var micontain = typeof(List<TValue>).GetMethod("Contains");
            Expression mc;
            Expression left = propertyName.Body;

            if (left.Type == typeof(long?) || left.Type == typeof(decimal?) || left.Type == typeof(DateTime?))
            {
                var valueExpression = Expression.Property(left, "Value");
                mc = Expression.Call(Expression.Constant(values), micontain, valueExpression);
            }
            else
            {
                mc = Expression.Call(Expression.Constant(values), micontain, left);
            }

            Expression body;
            switch (comparison)
            {
                case "NotIn":
                    body = Expression.Not(mc);
                    break;
                default:
                    body = mc;
                    break;
            }
            if (left.Type == typeof(long?) || left.Type == typeof(decimal?) || left.Type == typeof(DateTime?))
            {
                var hasValueExpression = Expression.Property(left, "HasValue");
                body = Expression.And(hasValueExpression, body);
            }

            return Expression.Lambda<Func<TSource, bool>>(body, propertyName.Parameters);
        }

        public static Expression<Func<TSource, bool>> BuildPredicate<TSource, TKey>(Expression<Func<TSource, TKey>> propertyName, string comparison, string value)
        {
            Expression left = propertyName.Body;
            Expression body, newLeft;
            newLeft = left;
            if (left.Type == typeof(string))
            {
                newLeft = Expression.Call(left, typeof(string).GetMethod("ToLower", System.Type.EmptyTypes));
            }

            switch (comparison)
            {
                case "NotContains":
                    body = Expression.Not(MakeComparison(newLeft, "Contains", value));
                    break;
                case "ReverseNotContains":
                    body = Expression.Not(MakeComparison(newLeft, "ReverseContains", value));
                    break;
                case "NotStartsWith":
                    body = Expression.Not(MakeComparison(newLeft, "StartsWith", value));
                    break;
                case "ReverseNotStartsWith":
                    body = Expression.Not(MakeComparison(newLeft, "ReverseStartsWith", value));
                    break;
                case "NotEndsWith":
                    body = Expression.Not(MakeComparison(newLeft, "EndsWith", value));
                    break;
                case "ReverseNotEndsWith":
                    body = Expression.Not(MakeComparison(newLeft, "ReverseEndsWith", value));
                    break;
                default:
                    body = MakeComparison(newLeft, comparison, value);
                    break;
            }
            if (left.Type == typeof(string) || left.Type == typeof(long?) || left.Type == typeof(decimal?) || left.Type == typeof(DateTime?))
            {
                var nullCheck = Expression.NotEqual(left, Expression.Constant(null, left.Type));
                body = Expression.And(nullCheck, body);
            }

            return Expression.Lambda<Func<TSource, bool>>(body, propertyName.Parameters);
        }

        public static Expression MakeComparison(Expression left, string comparison, string value)
        {
            switch (comparison)
            {
                case "==":
                    return MakeBinary(ExpressionType.Equal, left, value);
                case "!=":
                    return MakeBinary(ExpressionType.NotEqual, left, value);
                case ">":
                    return MakeBinary(ExpressionType.GreaterThan, left, value);
                case ">=":
                    return MakeBinary(ExpressionType.GreaterThanOrEqual, left, value);
                case "<":
                    return MakeBinary(ExpressionType.LessThan, left, value);
                case "<=":
                    return MakeBinary(ExpressionType.LessThanOrEqual, left, value);
                case "Contains":
                case "StartsWith":
                case "EndsWith":
                    return Expression.Call(MakeString(left), comparison, Type.EmptyTypes, Expression.Constant(value, typeof(string)));
                case "ReverseContains":
                    return Expression.Call(Expression.Constant(value, typeof(string)), "Contains", Type.EmptyTypes, MakeString(left));
                case "ReverseStartsWith":
                    return Expression.Call(Expression.Constant(value, typeof(string)), "StartsWith", Type.EmptyTypes, MakeString(left));
                case "ReverseEndsWith":
                    return Expression.Call(Expression.Constant(value, typeof(string)), "EndsWith", Type.EmptyTypes, MakeString(left));
                default:
                    throw new NotSupportedException($"Invalid comparison operator '{comparison}'.");
            }
        }

        public static Expression MakeString(Expression source)
        {
            return source.Type == typeof(string) ? source : Expression.Call(source, "ToString", Type.EmptyTypes);
        }

        public static Expression MakeBinary(ExpressionType type, Expression left, string value)
        {
            object typedValue = value;
            if (left.Type != typeof(string))
            {
                if (string.IsNullOrEmpty(value))
                {
                    typedValue = null;
                    if (Nullable.GetUnderlyingType(left.Type) == null)
                        left = Expression.Convert(left, typeof(Nullable<>).MakeGenericType(left.Type));
                }
                else
                {
                    var valueType = Nullable.GetUnderlyingType(left.Type) ?? left.Type;
                    typedValue = valueType.IsEnum ? Enum.Parse(valueType, value) :
                        valueType == typeof(Guid) ? Guid.Parse(value) :
                        Convert.ChangeType(value, valueType);
                }
            }
            var right = Expression.Constant(typedValue, left.Type);
            return Expression.MakeBinary(type, left, right);
        }
    }
}
