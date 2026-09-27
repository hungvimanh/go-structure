using Microsoft.EntityFrameworkCore.ChangeTracking.Internal;
using Microsoft.EntityFrameworkCore.Diagnostics;
using Microsoft.EntityFrameworkCore.Query;
using Microsoft.EntityFrameworkCore.Query.Internal;
using System.Data.Common;
using System.Reflection;
using Thinktecture;
using Thinktecture.EntityFrameworkCore.TempTables;

namespace CHECKIN.Common.FilterExtensions
{
    public static class SqlFilterExtension
    {
        private const long OptimizedElement = 200;
        private static DbContext GetDbContext(IQueryable query)
        {
            var bindingFlags = BindingFlags.NonPublic | BindingFlags.Instance;
            var queryCompiler = typeof(EntityQueryProvider).GetField("_queryCompiler", bindingFlags).GetValue(query.Provider);
            var queryContextFactory = queryCompiler.GetType().GetField("_queryContextFactory", bindingFlags).GetValue(queryCompiler);

            var dependenciesField = typeof(RelationalQueryContextFactory).GetField("<Dependencies>k__BackingField", bindingFlags);
            var dependencies = dependenciesField.GetValue(queryContextFactory);
            var queryContextDependencies = typeof(DbContext).Assembly.GetType(typeof(QueryContextDependencies).FullName);
            var stateManagerProperty = queryContextDependencies.GetProperty("StateManager", bindingFlags | BindingFlags.Public).GetValue(dependencies);
            var stateManager = (IStateManager)stateManagerProperty;

            return stateManager.Context;
        }

        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, string>> propertyName, StringFilter filter)
        {
            if (filter == null)
                return source;

            if (!string.IsNullOrEmpty(filter.Equal))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.ToLower()));

            if (!string.IsNullOrEmpty(filter.NotEqual))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.ToLower()));

            if (!string.IsNullOrEmpty(filter.Contain))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "Contains", filter.Contain.ToLower()));

            if (!string.IsNullOrEmpty(filter.NotContain))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "NotContains", filter.NotContain.ToLower()));

            if (!string.IsNullOrEmpty(filter.ReverseContain))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseContains", filter.ReverseContain.ToLower()));

            if (!string.IsNullOrEmpty(filter.ReverseNotContain))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseNotContains", filter.ReverseNotContain.ToLower()));

            if (!string.IsNullOrEmpty(filter.StartWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "StartsWith", filter.StartWith.ToLower()));

            if (!string.IsNullOrEmpty(filter.NotStartWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "NotStartsWith", filter.NotStartWith.ToLower()));

            if (!string.IsNullOrEmpty(filter.ReverseStartWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseStartsWith", filter.ReverseStartWith.ToLower()));

            if (!string.IsNullOrEmpty(filter.ReverseNotStartWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseNotStartsWith", filter.ReverseStartWith.ToLower()));

            if (!string.IsNullOrEmpty(filter.EndWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "EndsWith", filter.EndWith.ToLower()));

            if (!string.IsNullOrEmpty(filter.NotEndWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "NotEndsWith", filter.NotEndWith.ToLower()));

            if (!string.IsNullOrEmpty(filter.ReverseEndWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseEndsWith", filter.ReverseEndWith.ToLower()));

            if (!string.IsNullOrEmpty(filter.ReverseNotEndWith))
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseNotEndsWith", filter.ReverseEndWith.ToLower()));

            DbContext DbContext = GetDbContext(source);

            if (!string.IsNullOrEmpty(filter.CombineContain))
            {
                var source1 = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "Contains", filter.Contain.ToLower()));
                var source2 = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseContains", filter.Contain.ToLower()));
                source = source1.Union(source2);
            }

            if (!string.IsNullOrEmpty(filter.CombineStartWith))
            {
                var source1 = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "StartsWith", filter.ReverseStartWith.ToLower()));
                var source2 = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseStartsWith", filter.ReverseStartWith.ToLower()));
                source = source1.Union(source2);
            }

            if (!string.IsNullOrEmpty(filter.CombineEndWith))
            {
                var source1 = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "EndsWith", filter.ReverseEndWith.ToLower()));
                var source2 = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "ReverseEndsWith", filter.ReverseNotEndWith.ToLower()));
                source = source1.Union(source2);
            }

            if (filter.In != null)
            {
                if (filter.In.Any())
                {
                    ITempTableQuery<string> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
                else
                {
                    source = source.Where(x => false);
                }
            }

            if (filter.NotIn != null && filter.NotIn.Any())
            {
                ITempTableQuery<string> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, decimal>> propertyName, DecimalFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));
            if (filter.Less.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<", filter.Less.Value.ToString()));
            if (filter.LessEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<=", filter.LessEqual.Value.ToString()));
            if (filter.Greater.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">", filter.Greater.Value.ToString()));
            if (filter.GreaterEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">=", filter.GreaterEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Any())
                {
                    ITempTableQuery<decimal> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
                else
                {
                    source = source.Where(x => false);
                }
            }

            if (filter.NotIn != null && filter.NotIn.Any())
            {
                ITempTableQuery<decimal> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, decimal?>> propertyName, DecimalFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));
            if (filter.Less.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<", filter.Less.Value.ToString()));
            if (filter.LessEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<=", filter.LessEqual.Value.ToString()));
            if (filter.Greater.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">", filter.Greater.Value.ToString()));
            if (filter.GreaterEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">=", filter.GreaterEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Any())
                {
                    ITempTableQuery<decimal> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
                else
                {
                    source = source.Where(x => false);
                }
            }

            if (filter.NotIn != null && filter.NotIn.Any())
            {
                ITempTableQuery<decimal> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, long>> propertyName, LongFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));
            if (filter.Less.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<", filter.Less.Value.ToString()));
            if (filter.LessEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<=", filter.LessEqual.Value.ToString()));
            if (filter.Greater.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">", filter.Greater.Value.ToString()));
            if (filter.GreaterEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">=", filter.GreaterEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Any())
                {
                    ITempTableQuery<long> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
                else
                {
                    source = source.Where(x => false);
                }
            }

            if (filter.NotIn != null && filter.NotIn.Any())
            {
                ITempTableQuery<long> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, long?>> propertyName, LongFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));
            if (filter.Less.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<", filter.Less.Value.ToString()));
            if (filter.LessEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<=", filter.LessEqual.Value.ToString()));
            if (filter.Greater.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">", filter.Greater.Value.ToString()));
            if (filter.GreaterEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">=", filter.GreaterEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Any())
                {
                    ITempTableQuery<long> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
                else
                {
                    source = source.Where(x => false);
                }
            }

            if (filter.NotIn != null && filter.NotIn.Any())
            {
                ITempTableQuery<long> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, DateTime>> propertyName, DateFilter filter)
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));
            if (filter.Less.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<", filter.Less.Value.ToString()));
            if (filter.LessEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<=", filter.LessEqual.Value.ToString()));
            if (filter.Greater.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">", filter.Greater.Value.ToString()));
            if (filter.GreaterEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">=", filter.GreaterEqual.Value.ToString()));

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, DateTime?>> propertyName, DateFilter filter)
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));
            if (filter.Less.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<", filter.Less.Value.ToString()));
            if (filter.LessEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "<=", filter.LessEqual.Value.ToString()));
            if (filter.Greater.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">", filter.Greater.Value.ToString()));
            if (filter.GreaterEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, ">=", filter.GreaterEqual.Value.ToString()));

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, Guid>> propertyName, GuidFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Any())
                {
                    ITempTableQuery<Guid> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
                else
                {
                    source = source.Where(x => false);
                }
            }

            if (filter.NotIn != null && filter.NotIn.Any())
            {
                ITempTableQuery<Guid> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, Guid?>> propertyName, GuidFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));
            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Any())
                {
                    ITempTableQuery<Guid> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
                else
                {
                    source = source.Where(x => false);
                }
            }

            if (filter.NotIn != null && filter.NotIn.Any())
            {
                ITempTableQuery<Guid> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, long>> propertyName, IdFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));

            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Count == 0)
                    source = source.Where(x => false);
                else if (filter.In.Count < 200)
                    source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "In", filter.In));
                else
                {
                    ITempTableQuery<long> tempTableQuery = DbContext
                       .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
            }

            if (filter.NotIn != null)
            {
                if (filter.NotIn.Count == 0)
                    source = source.Where(x => true);
                else if (filter.NotIn.Count < 200)
                    source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "NotIn", filter.NotIn));
                else
                {
                    ITempTableQuery<long> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                    source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
                }
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, long?>> propertyName, IdFilter filter) where TSource : class
        {
            if (filter == null)
                return source;

            if (filter.Equal.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Equal.Value.ToString()));

            if (filter.NotEqual.HasValue)
                source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "!=", filter.NotEqual.Value.ToString()));

            DbContext DbContext = GetDbContext(source);

            if (filter.In != null)
            {
                if (filter.In.Count == 0)
                    source = source.Where(x => false);
                else if (filter.In.Count < 200)
                {
                    source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "In", filter.In));
                }
                else
                {
                    ITempTableQuery<long> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.In.Distinct().ToList()).GetAwaiter().GetResult();
                    source = source.Join(
                        tempTableQuery.Query,
                        propertyName,
                        x2 => x2,
                        (x1, x2) => x1);
                }
            }

            if (filter.NotIn != null)
            {
                if (filter.NotIn.Count == 0)
                    source = source.Where(x => true);
                else if (filter.NotIn.Count < 200)
                {
                    source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "NotIn", filter.NotIn));
                }
                else
                {
                    ITempTableQuery<long> tempTableQuery = DbContext
                   .BulkInsertValuesIntoTempTableAsync(filter.NotIn.Distinct().ToList()).GetAwaiter().GetResult();

                    source = source
                        .LeftJoin(
                            tempTableQuery.Query,
                            propertyName,
                            x2 => x2,
                            (outer, inner) => new { Left = outer, Right = inner })
                        .Where(x => x.Right == null)
                        .Select(x => x.Left);
                }
            }

            return source;
        }
        public static IQueryable<TSource> Where<TSource>(this IQueryable<TSource> source, Expression<Func<TSource, bool>> propertyName, bool? filter)
        {
            if (filter == null)
                return source;

            source = source.Where(ExpressionBuilder.BuildPredicate(propertyName, "==", filter.Value.ToString()));
            return source;
        }
       
    }

    public static class OrderByExtension
    {
        public static IQueryable<T> OrderBy<T, TOrderEnum>(this IQueryable<T> query,
            TOrderEnum orderBy,
            OrderType orderType,
            Dictionary<TOrderEnum, LambdaExpression> customOrder = null
        )
            where TOrderEnum : Enum
        {
            if (customOrder != null && customOrder.ContainsKey(orderBy))
            {
                query = query.OrderBy(customOrder[orderBy], orderType);
            }
            else
            {
                query = query.Order(orderBy.ToString(), orderType);
            }
            return query;
        }

        public static IQueryable<T> Paging<T, TFilter>(this IQueryable<T> query, TFilter filter)
            where TFilter : FilterEntity
        {
            query = query.Skip(filter.Skip);
            query = query.Take(filter.Take);
            return query;
        }

        private static IQueryable<TSource> Order<TSource>(this IQueryable<TSource> query, string propertyName, OrderType Type)
        {
            var order = Type == OrderType.ASC ? "OrderBy" : "OrderByDescending";
            var entityType = typeof(TSource);

            //Create x=>x.PropName
            var propertyInfo = entityType.GetProperty(propertyName);
            ParameterExpression arg = Expression.Parameter(entityType, "x");
            MemberExpression property = Expression.Property(arg, propertyName);
            var selector = Expression.Lambda(property, new ParameterExpression[] { arg });

            //Get System.Linq.Queryable.OrderBy() method.
            var enumarableType = typeof(Queryable);
            var method = enumarableType.GetMethods()
                 .Where(m => m.Name == order && m.IsGenericMethodDefinition)
                 .Where(m =>
                 {
                     var parameters = m.GetParameters().ToList();
                     //Put more restriction here to ensure selecting the right overload                
                     return parameters.Count == 2;//overload that has 2 parameters
                 }).Single();
            //The linq's OrderBy<TSource, TKey> has two generic types, which provided here
            MethodInfo genericMethod = method
                 .MakeGenericMethod(entityType, propertyInfo.PropertyType);

            /*Call query.OrderBy(selector), with query and selector: x=> x.PropName
              Note that we pass the selector as Expression to the method and we don't compile it.
              By doing so EF can extract "order by" columns and generate SQL for it.*/
            var newQuery = (IOrderedQueryable<TSource>)genericMethod
                 .Invoke(genericMethod, new object[] { query, selector });
            return newQuery;
        }

        private static IOrderedQueryable<TSource> OrderBy<TSource>(this IQueryable<TSource> query, LambdaExpression selector, OrderType Type)
        {
            var order = Type == OrderType.ASC ? "OrderBy" : "OrderByDescending";
            var entityType = typeof(TSource);

            //Get System.Linq.Queryable.OrderBy() method.
            var enumarableType = typeof(Queryable);
            var method = enumarableType.GetMethods()
                 .Where(m => m.Name == order && m.IsGenericMethodDefinition)
                 .Where(m =>
                 {
                     var parameters = m.GetParameters().ToList();
                     //Put more restriction here to ensure selecting the right overload                
                     return parameters.Count == 2;//overload that has 2 parameters
                 }).Single();
            //The linq's OrderBy<TSource, TKey> has two generic types, which provided here
            MethodInfo genericMethod = method
                 .MakeGenericMethod(entityType, selector.ReturnType);

            /*Call query.OrderBy(selector), with query and selector: x=> x.PropName
              Note that we pass the selector as Expression to the method and we don't compile it.
              By doing so EF can extract "order by" columns and generate SQL for it.*/
            var newQuery = (IOrderedQueryable<TSource>)genericMethod
                 .Invoke(genericMethod, new object[] { query, selector });
            return newQuery;
        }
    }

    public class HintCommandInterceptor : DbCommandInterceptor
    {
        public override InterceptionResult<DbDataReader> ReaderExecuting(
            DbCommand command,
            CommandEventData eventData,
            InterceptionResult<DbDataReader> result)
        {
            if (command.CommandText.Contains("INSERT") ||
                command.CommandText.Contains("UPDATE") ||
                command.CommandText.Contains("DELETE"))
                return result;
            command.CommandText = "SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;" + command.CommandText + " OPTION (FORCE ORDER)";
            return result;
        }

        public override ValueTask<InterceptionResult<DbDataReader>> ReaderExecutingAsync(DbCommand command, CommandEventData eventData, InterceptionResult<DbDataReader> result, CancellationToken cancellationToken = default)
        {
            if (command.CommandText.Contains("INSERT") ||
                 command.CommandText.Contains("UPDATE") ||
                 command.CommandText.Contains("DELETE"))
                return new ValueTask<InterceptionResult<DbDataReader>>(result);
            command.CommandText = "SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;" + command.CommandText + " OPTION (FORCE ORDER)";
            return new ValueTask<InterceptionResult<DbDataReader>>(result);
        }
    }
}
